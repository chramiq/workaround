# About

**Workaround routes your local traffic through cloudflare workers**, useful for:
* **WAF evasion / IP rotation:** leverage cloudflare's massive ip pool.
* **Traffic washing:** chain with TOR to bypass exit node filtering.

Heavily inspired by [flareprox](https://github.com/MrTurvey/flareprox)

## Installation

```bash
go install -v github.com/chramiq/workaround/cmd/workaround@latest
```

## Configuration

### First Run:
Run the tool once to generate the default configuration files.
```bash
workaround
```
*Creates `~/.config/workaround/config.json`*

### Getting Credentials

1. **Sign Up:** Create an account at [dash.cloudflare.com/sign-up](https://dash.cloudflare.com/sign-up).
2. **Generate Token:** Navigate to [API Tokens](https://dash.cloudflare.com/profile/api-tokens) and click **Create Token**.
3. **Template:** Select the **Edit Cloudflare Workers** template.
4. **Permissions:** 
    * Set **Account Resources** to `Include`  `All accounts`.
    * Set **Zone Resources** to `Include`  `All zones`.
5. **Finalize:** Click **Continue to Summary**  **Create Token**.
6. **Collect:** Copy your **API Token** and grab your **Account ID** from the main dashboard.

### Add Credentials:
Edit the config file and add your Cloudflare **Account ID** and **API Token**.

```json
{
  "accounts": [
    {
      "account_id": "YOUR_ACCOUNT_ID",
      "api_token": "YOUR_API_TOKEN"
    }
  ]
}
```

## Usage

### 1. Deploy and Verify
Build your worker fleet and verify connectivity.
```bash
workaround deploy
workaround status
workaround verify
```

### 2. Set Alias
Add the `wa` shortcut to your shell for seamless usage.
```bash
workaround alias
```

### 3. Run Commands
Prefix any network tool with `wa`.

**Basic Curl:**
```bash
wa curl http://ifconfig.me/ip
```

**Fuzzing (ffuf):**
```bash
wa ffuf -u http://target.com/FUZZ -w wordlist.txt
```
> **Note:** Always use `http://` in your commands. The local proxy accepts HTTP, and the Worker upgrades it to HTTPS automatically.

### 4. Manage
Check your Cloudflare free tier usage or cleanup resources.
```bash
workaround credits
workaround destroy
```

## Customization & Logs

All configuration and data are stored in `~/.config/workaround/`.

*   **Worker Script:** You can customize the Cloudflare Worker logic by editing `worker.js`. Run `workaround deploy` after changes to update your fleet.
*   **User-Agents:** Add or remove User-Agents in `useragents.txt`. The proxy picks one at random for each request (if `-r` is used).
*   **Logs:** Detailed activity and error logs are written to `debug.log`. Use this for troubleshooting connectivity issues.

## Advanced Usage

**Global Flags:**
| Flag | Description |
| :--- | :--- |
| `-v`, `--verbose` | Enable verbose console logs (silent by default). |

**Exec Flags (pass to `wa`):**
| Flag | Description |
| :--- | :--- |
| `-r`, `--random-useragent` | Enable User-Agent randomization (Default: Transparent). |
| `-c`, `--new-circuit` | Force a new Tor exit node for this session. |
| `-u`, `--unsafe-http` | Disable the auto-downgrade safety check for URLs. |
| `--http` | Force the Worker to use HTTP (do not upgrade to HTTPS). |

> **Tip:** You can combine single-letter flags!
> Example: `wa -rc curl ...` (Random UA + New Circuit)

### Tor Integration (Traffic Washing)
To route traffic through Tor before hitting Cloudflare:

1.  Create Cloudflare account **via Tor**, get credentials.
2.  Ensure Tor is running (`sudo service tor start`).
3.  Edit `~/.config/workaround/config.json`:
  ```json
  "upstream_proxy": "socks5://127.0.0.1:9050"
  ```
> **Flow:** You $\to$ Local Proxy $\to$ Tor $\to$ Cloudflare Worker $\to$ Target.

### Debugging
To see exactly what the worker constructs before sending it to the target:
```bash
wa curl http://httpbin.org/headers
```
