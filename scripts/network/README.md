# Optional outbound Tor setup (macOS)

This helper runs a separate, outbound-only Tor service on `127.0.0.1:9052`.
It starts after login and restarts after a process exit. It cannot keep a sleeping
or shut-down Mac online. Existing inbound Tor services and identities remain in place.
Requires an installed Tor executable and Python 3. Install Tor from its official
distribution or Homebrew (`brew install tor`).

Run `python3 outbound-tor.py install` to install the optional service.
If you already have this Grynvault service, use `python3 outbound-tor.py status`;
the installer refuses to overwrite an existing configuration.
Wait for `Bootstrapped 100%` in
`~/Library/Application Support/Grynvault Network/service.log` before using it.
To stop and remove this service, run `python3 outbound-tor.py uninstall`.
Configuration and Tor data are retained. Reinstallation after removal requires
reviewing those retained files; the helper will not overwrite them.

## Opt-in peer routing

The app automatically uses its improved reconnection policy. To use native Tor
transport for a particular peer, obtain its verified public key and v3 onion
address from that peer. Quit Grynvault normally before adding a
`peer_transport_overrides.json` file inside the app's existing `ldk/storage`
directory. This is a separate JSON configuration file; do not edit the database.
For the sandboxed macOS app the directory is:

`~/Library/Containers/com.grynvault.rebalance/Data/Library/Application Support/grynvault-rebalance/ldk/storage/`

The format is an object mapping each peer public key to:

```json
{
  "PEER_PUBLIC_KEY": {
    "proxy": "127.0.0.1:9052",
    "host": "PEER_V3_ONION_ADDRESS.onion",
    "port": 9735
  }
}
```

Replace both placeholders before saving; they are deliberately invalid examples.
Only loopback proxies and v3 onion-shaped destinations are accepted. Lightning
still authenticates the expected peer public key. Invalid configuration prevents
node startup; correct or remove only this JSON file to return to normal routing.
Do not copy, restore, or replace an older channel database during rollback.

Tor setup completes before the Lightning handshake timer begins. Failed attempts
use 30, 60, 120, then 240 second delays, checked on a 60 second polling cycle.
Each peer connects independently. Remote outages, liquidity, invoice expiry and
routing limits can still cause payment failures. No payment is sent by this helper.

For an offline installer check, use `python3 outbound-tor.py prepare --directory
/path/to/new/test-folder`; it validates the configuration without starting Tor.
