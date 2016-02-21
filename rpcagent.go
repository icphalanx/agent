package agent

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"

	"fmt"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/icphalanx/agent/types"
	pb "github.com/icphalanx/rpc"
)

import (
	"log"
)

var (
	ErrExitingForCertRotation = fmt.Errorf(`exiting to allow certificate rotation`)
)

type RPCAgent struct {
	client pb.PhalanxCollectorClient
	agent  types.Host
	conn   *grpc.ClientConn
	cert   *tls.Certificate

	logLineChan chan types.ReporterLogLine
}

func (r *RPCAgent) init() error {
	var err error

	rpcAgent, err := types.HostToRPC(r.agent)
	if err != nil {
		return err
	}

	_, err = r.client.ConfigureMe(context.TODO(), rpcAgent)

	return err
}

func (r *RPCAgent) certDueForRenewal() bool {
	if r.cert.Leaf == nil {
		var err error
		r.cert.Leaf, err = x509.ParseCertificate(r.cert.Certificate[0])
		if err != nil {
			log.Println("got error whilst checking if cert is due for renewal", err)

			// this will cause getNewCertificateIfNeeded to f
			// which should hopefully propagate the same error
			return true
		}
	}

	return r.cert.Leaf.NotAfter.Before(time.Now().Add(30 * 24 * time.Hour))
}

func (r *RPCAgent) getNewCertificateIfNeeded(caCertPool *x509.CertPool, certPath, privKeyPath string) (*tls.Config, error) {
	var err error
	r.cert.Leaf, err = x509.ParseCertificate(r.cert.Certificate[0])
	if err != nil {
		return nil, err
	}

	if r.cert.Leaf.Subject.CommonName != `Phalanx Provisioning Certificate` || r.certDueForRenewal() {
		// we don't need to get a new certificate
		return nil, nil
	}

	commonName, err := r.agent.HumanName()
	if err != nil {
		return nil, err
	}

	// open our file handles
	certFile, err := os.Create(certPath)
	if err != nil {
		return nil, err
	}
	defer certFile.Close()

	privKeyFile, err := os.Create(privKeyPath)
	if err != nil {
		return nil, err
	}
	defer privKeyFile.Close()

	// we need a new certificate
	// so... generate a new keypair
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	privAsn := x509.MarshalPKCS1PrivateKey(priv)
	privPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privAsn,
	})
	if privKeyFile.Write(privPem); err != nil {
		return nil, err
	}

	csr := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: commonName,
		},
		DNSNames: []string{commonName},
	}
	genCsr, err := x509.CreateCertificateRequest(rand.Reader, &csr, priv)
	if err != nil {
		return nil, err
	}

	csrPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: genCsr,
	})

	signingResp, err := r.client.SignMe(context.TODO(), &pb.SigningRequest{
		Csr: string(csrPem),
	})
	if err != nil {
		return nil, err
	}

	if _, err := certFile.WriteString(signingResp.Cert); err != nil {
		return nil, err
	}

	kp, err := tls.X509KeyPair([]byte(signingResp.Cert), privPem)
	if err != nil {
		return nil, err
	}

	r.cert = &kp
	return generateTLSConfig(caCertPool, kp)
}

func (r *RPCAgent) logLineHandler() {
	log.Println("loglinehandler: starting up")
	stream, err := r.client.RecordLogs(context.Background())
	if err != nil {
		log.Fatalf("loglinehandler: failed to RecordLogs: %v", err)
	}

	for {
		lc := <-r.logLineChan
		hn, err := lc.Host.HumanName()
		if err != nil {
			log.Println("loglinehandler: failed to get HumanName for %s", lc)
			continue
		}
		if err := stream.Send(&pb.LogLine{
			Reporter:  lc.Reporter.Id(),
			Timestamp: types.TimeToGoogleTimestamp(lc.Timestamp),
			Line:      lc.LogLine,
			Host:      hn,
			Tags:      lc.Tags,
		}); err != nil {
			log.Fatalf("loglinehandler: failed to stream.Send: %v", err)
		}
	}
}

func (r *RPCAgent) Run() error {
	// do a first run
	err := r.tick()
	if err != nil {
		return err
	}

	exitCh := make(chan struct{})
	go func(exitCh chan<- struct{}) {
		sleepFor := r.cert.Leaf.NotAfter.Sub(time.Now()) - (20 * 24 * time.Hour)
		log.Println("certrenew: sleeping for", sleepFor)
		time.Sleep(sleepFor)
		log.Println("certrenew: woke up!")
		exitCh <- struct{}{}
	}(exitCh)

	// log line chan
	go r.logLineHandler()
	// and spin up handlers
	reporters, err := r.agent.Reporters()
	if err != nil {
		return err
	}
	for _, reporter := range reporters {
		go func(reporter types.Reporter) {
			llc := reporter.LogLines()
			for {
				r.logLineChan <- <-llc
			}
		}(reporter)
	}

	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-exitCh:
			close(r.logLineChan)
			return ErrExitingForCertRotation
		case <-ticker.C:
			err := r.tick()
			if err != nil {
				close(r.logLineChan)
				return err
			}
		}
	}
}

func (r *RPCAgent) tick() error {
	var err error
	log.Println("tick...")
	rep := new(pb.ReportRequest)

	rep.Host, err = types.HostToRPC(r.agent)
	if err != nil {
		return err
	}

	reporters, err := r.agent.Reporters()
	if err != nil {
		return err
	}

	rep.Reporters, err = types.ReportersToRPC(reporters)
	if err != nil {
		return err
	}

	resp, err := r.client.Report(context.TODO(), rep)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("remote reported !success")
	}
	return nil
}

func generateCertPoolFromPath(caPath string) (*x509.CertPool, error) {
	b, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, err
	}

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(b) {
		return nil, fmt.Errorf("failed to load certificate from PEM")
	}
	return cp, nil
}

func generateTLSConfig(caCertPool *x509.CertPool, cert tls.Certificate) (*tls.Config, error) {
	c := new(tls.Config)
	c.RootCAs = caCertPool
	c.Certificates = []tls.Certificate{cert}
	return c, nil
}

func rpcAgentWithConfig(target string, agent types.Host, tlsConfig *tls.Config) (*RPCAgent, error) {
	cert := tlsConfig.Certificates[0]
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		return nil, err
	}

	client := pb.NewPhalanxCollectorClient(conn)
	r := RPCAgent{
		agent:       agent,
		client:      client,
		conn:        conn,
		cert:        &cert,
		logLineChan: make(chan types.ReporterLogLine, 10),
	}

	return &r, nil
}

func NewRPCAgent(target string, caPath, certPath, privKeyPath string) (*RPCAgent, error) {
	// if caPath doesn't exist, abort
	caCertPool, err := generateCertPoolFromPath(caPath)
	if err != nil {
		return nil, err
	}

	// load the certificate
	var cert tls.Certificate
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		// we need to get the provisioning certificate from the binary
		cert, err = LoadEmbeddedCertPair()
	} else {
		cert, err = tls.LoadX509KeyPair(certPath, privKeyPath)
	}
	if err != nil {
		return nil, err
	}

	// make the TLS config
	tlsConfig, err := generateTLSConfig(caCertPool, cert)
	if err != nil {
		return nil, err
	}

	agent, err := MakeLocalAgent()
	if err != nil {
		return nil, err
	}

	r, err := rpcAgentWithConfig(target, agent, tlsConfig)
	if err != nil {
		return nil, err
	}

	tlsConfig, err = r.getNewCertificateIfNeeded(caCertPool, certPath, privKeyPath)
	if err != nil {
		return nil, err
	}

	if tlsConfig != nil {
		// OK, swap out the cert for the new ones we just wrote
		if err := r.conn.Close(); err != nil {
			return nil, err
		}

		r, err = rpcAgentWithConfig(target, agent, tlsConfig)
		if err != nil {
			return nil, err
		}
	}

	return r, r.init()
}
