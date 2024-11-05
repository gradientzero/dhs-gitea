## Copy dhs-gitea server
The purpose of this scripts are to copy dhs-gitea server to another dhs-gitea server.

## Requirements
- Live origin dhs-gitea server
- Live destination dhs-gitea server

## Scripts Breakdown
- `autorun.sh` - Run other 3 scripts and clean files.
- `1-old-server.sh` - Get all required data from origin server, run from origin server.
- `2-host-server.sh` - Transfer data from origin server to destination server, run from host machine.
- `3-new-server.sh` - Put data into destination server.

## Usage
`autorun.sh` will automatically execute the 3 other scripts. Run it on host machine along with 6 parameters as arguments.
Then all unneeded files will be cleaned as well.

```sh
./autorun.sh <origin_server> <origin_docker_container> <origin_tenant> <destination_server> <destination_docker_container> <destination_tenant>
```

- First arg: origin_server, ex: root@0.0.0.0
- Second arg: origin_docker_container, ex: 5f651399aab2
- Third arg: origin_tenant, ex: usb
- Fourth arg: destination_server, ex: root@1.1.1.1
- Fifth arg: destination_docker_container, ex: bfd6e62fa2d0
- Sixth arg: destination_tenant, ex: dhcs

## Example:
```sh
./autorun.sh root@0.0.0.0 5f651399aab2 usb root@1.1.1.1 bfd6e62fa2d0 dhcs
```