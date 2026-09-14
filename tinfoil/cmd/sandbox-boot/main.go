package main

import (
	"tinfoil/boot"
	"tinfoil/internal/attestation"
	variant "tinfoil/internal/sandboxvariant"
)

func bootSpec() boot.Spec {
	return boot.Spec{
		Stages:         variant.BootStages(),
		Validate:       validate,
		AttestDevices:  boot.InferenceSpec().AttestDevices,
		IsolateModel:   func(*boot.Config, string) bool { return true },
		DeviceEvidence: attestation.CollectDeviceEvidence,
	}
}

func main() { boot.Main(bootSpec()) }

func validate(config *boot.Config) error { return variant.Validate(config) }
