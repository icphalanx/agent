package packagekit

import (
	"fmt"
	"github.com/godbus/dbus"
	"github.com/icphalanx/agent/types"
	"log"
)

type PackageKitReporter struct {
	dbusConn *dbus.Conn
	dbusObj  dbus.BusObject
}

func (PackageKitReporter) Id() string {
	return "packagekit"
}

func (PackageKitReporter) Issues() ([]types.Issue, error) {
	return []types.Issue{}, nil
}

func (pkr PackageKitReporter) createTransaction(cb func(dbus.BusObject, <-chan *dbus.Signal) error) error {
	call := pkr.dbusObj.Call("org.freedesktop.PackageKit.CreateTransaction", 0)
	if call.Err != nil {
		return call.Err
	}

	var dbusPath dbus.ObjectPath
	if err := call.Store(&dbusPath); err != nil {
		return err
	}

	matchRule := fmt.Sprintf(
		"type='signal',path='%s',interface='org.freedesktop.PackageKit.Transaction'",
		dbusPath,
	)

	defer pkr.dbusConn.BusObject().Call("org.freedesktop.DBus.RemoveMatch", 0, matchRule)
	pkr.dbusConn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, matchRule)

	ch := make(chan *dbus.Signal, 10)
	pkr.dbusConn.Signal(ch)
	defer pkr.dbusConn.Signal(ch) // unregister once we're done

	trans := pkr.dbusConn.Object("org.freedesktop.PackageKit", dbusPath)

	return cb(trans, ch)
}

func (pkr PackageKitReporter) countPackages(txCall string, filter PackageKitFilterBitField) (uint, error) {
	count := uint(0)
	txCall = fmt.Sprintf("org.freedesktop.PackageKit.Transaction.%s", txCall)
	err := pkr.createTransaction(func(trans dbus.BusObject, ch <-chan *dbus.Signal) error {
		err := trans.Call(txCall, 0, filter).Err
		if err != nil {
			return err
		}

		for {
			s := <-ch
			if s.Name == "org.freedesktop.PackageKit.Transaction.Finished" || s.Name == "org.freedesktop.PackageKit.Transaction.Destroy" {
				break
			}
			if s.Name == "org.freedesktop.PackageKit.Transaction.Package" {
				count += 1
			}
		}

		return nil
	})
	return count, err
}

func (pkr PackageKitReporter) Metrics() ([]types.Metric, error) {
	metrics := []types.Metric{}

	installedFilter := PackageKitFilterBitField(PK_FILTER_ENUM_INSTALLED)

	mych := make(chan *dbus.Signal, 10)
	pkr.dbusConn.Signal(mych)
	defer pkr.dbusConn.Signal(mych) // unregister
	go func() {
		for {
			s := <-mych
			log.Println("RECV", s)
		}
	}()

	// fetch number of packages needing updates
	if needUpdatePackages, err := pkr.countPackages("GetUpdates", 0); err == nil {
		metrics = append(metrics, PackageCountMetric{
			id: "needupdate",

			humanName: "Packages requiring updates",
			humanDesc: "The number of packages for which updates are available in the configured enabled software repositories.",

			packageCount: needUpdatePackages,
			shouldWarn:   true,
			warningLevel: 1,
			dangerLevel:  21,
		})
	}

	// fetch installed packages count
	if installedPackages, err := pkr.countPackages("GetPackages", installedFilter); err == nil {
		metrics = append(metrics, PackageCountMetric{
			id: "installed",

			humanName: "Installed packages",
			humanDesc: "The number of packages installed on the system.",

			packageCount: installedPackages,
		})
	}

	// fetch enabled repos
	{
		repos := []string{}
		err := pkr.createTransaction(func(trans dbus.BusObject, ch <-chan *dbus.Signal) error {
			err := trans.Call("org.freedesktop.PackageKit.Transaction.GetRepoList", 0, uint64(0)).Err
			if err != nil {
				return err
			}

			for {
				s := <-ch
				if s.Name == "org.freedesktop.PackageKit.Transaction.Finished" || s.Name == "org.freedesktop.PackageKit.Transaction.Destroy" {
					break
				}
				if s.Name == "org.freedesktop.PackageKit.Transaction.RepoDetail" {
					repoName := s.Body[0].(string)
					repoEnabled := s.Body[2].(bool)
					if repoEnabled {
						repos = append(repos, repoName)
					}
				}

			}

			return nil
		})
		if err == nil {
			metrics = append(metrics, RepoListMetric{repos})
		}
	}

	return metrics, nil
}

func (PackageKitReporter) Hosts() ([]types.Host, error) {
	return []types.Host{}, nil
}
