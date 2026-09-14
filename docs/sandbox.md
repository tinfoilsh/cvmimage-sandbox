# Sandbox image

Build the sandbox with `nix-build -I . -A shipping-image -o result`. The result
contains `tinfoilcvm.raw`, `.vmlinuz`, `.initrd`, and `.roothash`; `debug-image`
adds the measured debug console. Both are the `sandbox.*` derivations; the
inference variant is not exposed by this repository.

The sandbox keeps the boot, shim, PID 1, and volume code in the `tinfoil` Go module.
It omits Docker and containerd from its rootfs and keeps the NVIDIA modules,
programs, and nvattest, so a workload declaring `gpus` attests and uses them as
the inference image does. Its additional runtime is OpenSSH. Enrollment calls
`internal/volume.OpenWorkspace` directly. A future storage service can replace
that call without changing `/enroll`, SSH, or the workspace paths.

Use `sandbox/tinfoil-config.yml` as the measured workload definition. It pins an
executable toolchain pack and declares one root-owned executable volume named
`workspace`. The first storage disk follows the model disks in the existing
fixed PCI layout. The volume has no `key-secret`; the client supplies its key at
enrollment. The launcher must attach the workspace before enrollment. A zeroed
disk is formatted; an existing disk must authenticate with the supplied key.

The toolchain pack supplies `nix/store` and `nix/var/nix/profiles/default`.
The shared volume library mounts a writable store overlay and exports the same
tree at `/workspace` and `/nix`. The SSH home is `/workspace/home`. The image's
shell links resolve through `/nix/var/nix/profiles/default/bin` after enrollment.

A client first verifies the enclave's release measurement and reads `/healthz`
over its attested channel. The response supplies a boot nonce and the ephemeral
SSH host-key fingerprint. The orchestrator signs an ES256 permit with issuer
`https://orchestrator.tinfoil.sh`, the sandbox domain as subject, that nonce as
audience, and an expiration. Its verifying key is compiled into the image.

`POST /enroll` takes `Authorization: Bearer <permit>` and a JSON object with
`key`, one SSH public-key line, and `volume`, base64 for the 64-byte workspace key.
After the authenticated disk, workspace layout, and exports are ready, the
volume library extends RTMR3 with the identity derived from the workspace key,
the same seal runtime unlock uses and the CLI recomputes from its `disk.key`.
Enrollment claims that owner and starts OpenSSH. Success returns 204;
a repeat returns 409. A failed unlock leaves ownership unclaimed. If SSH startup
fails after sealing, enrollment returns 503 and keeps the owner claimed.

SSH accepts only the enrolled public key and binds to loopback port 22. Connect
through the shim's HTTP/2 CONNECT endpoint with authority `localhost:22`, checking
the SSH fingerprint against `/healthz`. The shim also allows the fixed development
ports in `tinfoil/cmd/sandbox-shim/main.go`. A sandbox process exit fails the boot;
PID 1 does not restart enrollment within the same boot.

After reboot, enroll again using a fresh permit and the same workspace key. The
nonce and SSH host key change; workspace files and store-overlay writes persist.
On TDX, verify the seal in RTMR3 in a fresh report. SNP and ordinary VMs do
not provide this register. This image retains the existing authenticated-volume
format; it does not migrate older, unauthenticated sandbox disks.

Run `nix-build -A checks` for unit tests, race checks, debug PID 1 tests, and vet.

Release CI publishes the sandbox under the standard `tinfoil-inference-<version>`
artifact names: manifest, checksums, and provenance on the GitHub release, and
the disk, kernel, and initrd at `https://images.tinfoil.sh/cvm-sandbox/`. A
workload selects a release by adding `cvm-source` beside its `cvm-version`, as
`sandbox/tinfoil-config.yml` shows.
