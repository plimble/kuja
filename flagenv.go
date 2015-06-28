package kuja

import (
	"github.com/plimble/flagenv"
	"time"
)

type Config struct {
	Addr           string
	ShutdownTimout time.Duration
}

func FlaagEnv(fe *flagenv.FlagEnv) {
	if fe == nil {
		return
	}

	fe.AddString("addr", ":3000", "Listten address")
	fe.AddDuration("shutdown-timeout", 0, "Timeout on shutdown")
}
