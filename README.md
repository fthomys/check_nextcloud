# Nextcloud Icinga Plugin

This Go-based plugin checks the health of your Nextcloud instance by querying its API and evaluating system metrics (disk usage, CPU load, memory usage, and swap usage). It is designed to be used with Icinga and returns standard exit codes:

- **0 (OK)**
- **1 (WARNING)**
- **2 (CRITICAL)**

## Features

- **Nextcloud API Check:** Retrieves system, storage, and server details from Nextcloud.
- **System Metrics:** Uses local system calls to assess disk usage.
- **Thresholds:** Compares metrics (free disk space, CPU load, memory usage, and swap usage) against configurable warning and critical thresholds.
- **Performance Data:** Outputs key metrics in a format that Icinga can ingest.

## Requirements

- [Go](https://golang.org/doc/install) (version 1.20 or later recommended)
- A running Nextcloud instance with API enabled.
- A valid NC-Token for API authentication.
- Icinga 2 (or a compatible monitoring system such as Nagios).

## Installation

### Building the Plugin

1. **Clone the Repository:**

```bash
git clone https://github.com/fthomys/check_nextcloud.git
cd check_nextcloud
```

2. **Build the Executable:**

```bash
go build -o check_nextcloud main.go
```
or 
```bash
make
```

3. **Install the Plugin:**

Copy the executable to your Icinga (or Nagios) plugins directory. For example:

```bash
sudo cp check_nextcloud /usr/local/lib/nagios/plugins/
sudo chmod +x /usr/local/lib/nagios/plugins/check_nextcloud
```

> **Note:** Adjust the target directory according to your environment (e.g., `/usr/lib/nagios/plugins/`).

## Usage

You can run the plugin directly from the command line to test it:

```bash
/usr/local/lib/nagios/plugins/check_nextcloud -s https://your-nextcloud-url -t your_nc_token -c 5 -w 25 -d /
```

### Command-Line Options

| Option | Description |
|--------|-------------|
| `-s, --server` | Nextcloud Server URL (e.g., `https://cloud.example.com`) |
| `-t, --token` | Nextcloud NC-Token for authentication |
| `-c, --critical` | Critical threshold for free disk space in percent (default: 5) |
| `-w, --warning` | Warning threshold for free disk space in percent (default: 25) |
| `-d, --disklocation` | Disk location to check (default: `/`) |

## Icinga Configuration

To integrate this plugin with Icinga 2, you need to define a command object and a service. Below are example configuration snippets.

### Command Definition

Create a file (or add to an existing commands configuration file) with the following content:

```icinga2
object CheckCommand "check_nextcloud" {
    import "plugin-check-command"
    command = [ PluginDir + "/check_nextcloud" ]
    arguments = {
        "-s" = "$nextcloud_server$"
        "-t" = "$nextcloud_token$"
        "-c" = "$nextcloud_critical$"
        "-w" = "$nextcloud_warning$"
        "-d" = "$nextcloud_disklocation$"
    }
}
```

> **Note:** Replace `PluginDir` with the path where your plugins reside (e.g., `/usr/local/lib/nagios/plugins`). Define the custom variables (`nextcloud_server`, `nextcloud_token`, etc.) in your host or service definitions.

### Service Definition

Add a service definition for your Nextcloud host. For example:

```icinga2
apply Service "nextcloud-api-health" {
    import "generic-service"
    
    check_command = "check_nextcloud"
    
    assign where host.name == NodeName
    
    vars.nextcloud_server = "https://cloud.example.com"
    vars.nextcloud_token = "your_nc_token"
    vars.nextcloud_critical = 5
    vars.nextcloud_warning  = 25
    vars.nextcloud_disklocation = "/"
}
```

> **Reminder:** Replace `"https://cloud.example.com"`, and `"your_nc_token"` with the actual URL, and token. Adjust threshold values and disk location as needed.

## Testing the Plugin

Before deploying in production, test the plugin manually:

```bash
/usr/local/lib/nagios/plugins/check_nextcloud -s https://your-nextcloud-url -t your_nc_token -c 5 -w 25 -d /
```

You should see an output similar to:

```
OK - Nextcloud 30.0.4.1 running. | version=30.0.4.1 num_users=12 num_files=1971 free_space_bytes=894427783168 free_space_percent=75 cpu_load_1m=0.57421875 cpu_load_5m=0.3876953125 cpu_load_15m=0.353515625 memory_total=65643520 memory_free=54658048 memory_usage_percent=16 swap_total=33519616 swap_free=33519616 swap_usage_percent=0 num_apps_installed=50 num_apps_update_available=4 num_shares=0 php_version=8.2.27 db_version=11.4.4 active_users_5m=1 opcache_hit_rate=96.2478999439985
```