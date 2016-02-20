package packagekit

import (
	"github.com/godbus/dbus"
	"github.com/icphalanx/agent/reporters"
	"github.com/icphalanx/agent/types"
)

func init() {
	reporters.Register(PackageKitReporterFactory{})
}

type PackageKitReporterFactory struct{}

func (PackageKitReporterFactory) Id() string {
	return "packagekit"
}

func (pkrf PackageKitReporterFactory) Create(h types.Host) (types.Reporter, error) {
	if at, err := pkrf.ApplicableTo(h); !at {
		return nil, err
	}

	dbusConn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	return PackageKitReporter{
		dbusConn: dbusConn,
		dbusObj:  dbusConn.Object("org.freedesktop.PackageKit", "/org/freedesktop/PackageKit"),
	}, nil
}

func (PackageKitReporterFactory) ApplicableTo(h types.Host) (bool, error) {
	// is this a LinuxHost?
	if !h.IsLocal() {
		return false, nil
	}

	// does dbus work?
	conn, err := dbus.SystemBus()
	if err != nil {
		return false, err
	}

	// can we communicate with PackageKit?
	obj := conn.Object("org.freedesktop.PackageKit", "/org/freedesktop/PackageKit")
	_, err = obj.GetProperty("org.freedesktop.PackageKit.VersionMajor")
	if err != nil {
		return false, err
	}

	return true, err
}
