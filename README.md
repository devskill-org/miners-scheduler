# Miners Scheduler

A Go application that automatically manages Avalon miners based on electricity price data from the ENTSO-E Transparency Platform API. The scheduler periodically checks electricity prices and controls miners to optimize mining operations during low-cost periods.

## Features

- **Automatic Miner Discovery**: Scans your network to find Avalon miners
- **Real-time Price Monitoring**: Fetches electricity prices from ENTSO-E API
- **Smart Miner Management**: Automatically starts/stops miners based on configurable price thresholds
- **Thermal Protection**: Monitors fan speeds and automatically switches miners to Eco mode when overheating
- **Dry-run Mode**: Test your configuration and see what actions would be taken without executing them
- **Concurrent Operations**: Manages multiple miners simultaneously for efficiency
- **Robust Error Handling**: Includes retry logic and comprehensive logging
- **Flexible Configuration**: Support for both command-line arguments and configuration files
- **Health Monitoring**: Optional health check endpoint for monitoring

## How It Works

The scheduler performs two main tasks:

### Price-Based Management (every 15 minutes, configurable)
1. **Discover Miners**: Scans the specified network to find available Avalon miners
2. **Check Current Price**: Attempts to get the current electricity price from cached data
3. **Download New Data**: If price not found in cache, downloads latest market document from ENTSO-E API
4. **Price Comparison**: Compares current price with the configured limit
5. **Miner Management**:
   - If `price ≤ limit`: Wake up miners that are in standby mode
   - If `price > limit`: Put active miners into standby mode

### Thermal Protection (every 1 minute, configurable)
1. **State Monitoring**: Checks the current state of all discovered miners
2. **Fan Speed Analysis**: Monitors FanR (fan ratio) percentage for each miner
3. **Thermal Protection**: If FanR > 70%, automatically switches miner to Eco work mode
4. **Recovery**: When FanR ≤ 60%, switches miner back to Standard work mode

## Prerequisites

- Go 1.25.1 or later
- ENTSO-E Transparency Platform API token (get it from [ENTSO-E TP](https://transparency.entsoe.eu/))
- Network access to Avalon miners
- Avalon miners with API access enabled

## Installation

### From Source

```bash
git clone https://github.com/devskill-org/miners-scheduler.git
cd miners-scheduler
go build -o miners-scheduler ./
```

### Using Docker

```bash
docker build -t miners-scheduler .
docker run miners-scheduler
```

## Configuration

### Command Line Usage

```bash
# Basic usage with defaults
./miners-scheduler

# Custom configuration
./miners-scheduler --config "custom.json"


# Show help
./miners-scheduler -help
```

### Configuration File

Create a `config.json` file for more advanced configuration:

```json
{
  "price_limit": 50.0,
  "network": "192.168.1.0/24",
  "check_price_interval": "15m",
  "miners_state_check_interval": "1m",
  "dry_run": false,
  "api_timeout": "30s",
  "log_level": "info",
  "log_format": "text",
  "miner_timeout": "5s",
  "health_check_port": 8080,
  "security_token": "<security token>",
  "fanr_high_threshold": 70,
  "fanr_low_threshold": 50,
  "url_format": "https://web-api.tp.entsoe.eu/api?documentType=A44&out_Domain=10YLV-1001A00074&in_Domain=10YLV-1001A00074&periodStart=%s&periodEnd=%s&securityToken=%s",
  "location": "CET"
}
```

## Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `price_limit` | 50.0 | Price threshold in EUR/MWh |
| `network` | "192.168.1.0/24" | Network to scan for miners (CIDR notation) |
| `check_price_interval` | 15m | How often to run the scheduler task for price checking |
| `miners_state_check_interval` | 1m | How often to check miner states and manage thermal protection |
| `dry_run` | false | Run in dry-run mode (simulate actions without executing) |
| `api_timeout` | 30s | Timeout for ENTSO-E API calls |
| `log_level` | info | Logging level (debug, info, warn, error) |
| `log_format` | text | Log format (text, json) |
| `miner_timeout` | 5s | Timeout for miner operations |
| `health_check_port` | 0 | Health check port (0 = disabled) |
| `security_token` | <token> | ENTSO-E API Token |
| `fanr_high_threshold` | 70 | Fan rotation percentage when miner mode decreased to prevent overheating |
| `fanr_low_threshold` | 50 | Fan rotation percentage when miner mode increased |
| `url_format` | <Latvia prices URL> | ENTSO-E API URL Format to query electricity prices |
| `location` | CET | Timezone for daily prices publish period from 00:00 to 00:00 (next day) |


## Usage Examples

### Basic Setup

1. Get your ENTSO-E API token from the [transparency platform](https://transparency.entsoe.eu/)
2. Put token to `config.json`
3. Run the scheduler:
   ```bash
   ./miners-scheduler
   ```

### Testing Configuration with Dry-run Mode

Before running the scheduler in production, test your configuration with dry-run mode.

In dry-run mode, the scheduler will:
- Discover miners on your network
- Fetch current electricity prices
- Log what actions it would take
- **NOT** actually control the miners

Example dry-run output:
```
[SCHEDULER] DRY-RUN MODE ENABLED: Actions will be simulated only
[SCHEDULER] Current electricity price: 65.50 EUR/MWh
[SCHEDULER] Price limit: 60.00 EUR/MWh
[SCHEDULER] DRY-RUN: Would put miner 192.168.1.100:4028 into standby (price 65.50 > limit 60.00)
[SCHEDULER] DRY-RUN: Successfully simulated management of 3 miners
```



## API Reference

### Avalon Miner States

The scheduler recognizes three miner states:

- `Running` (0): Miner is active and running
- `Mining` (1): Miner is actively mining
- `StandBy` (2): Miner is in standby/idle mode

### Avalon Work Modes

The scheduler can manage three work modes for thermal protection:

- `Eco Mode` (0): Reduced power consumption and heat generation
- `Standard Mode` (1): Normal operation mode (default)
- `Super Mode` (2): Maximum performance mode

### Scheduler Operations

- **WakeUp**: Transitions miner from standby to active state
- **Standby**: Transitions miner from active to standby state
- **SetWorkMode**: Changes miner work mode for thermal management
- **GetLiteStats**: Retrieves current miner status and statistics


## Troubleshooting

### Common Issues

1. **No miners discovered**
   - Check network configuration (CIDR notation)
   - Ensure miners are accessible and have API enabled
   - Verify firewall settings

2. **API errors**
   - Verify API token is correct
   - Check internet connectivity
   - Ensure ENTSO-E API is accessible

3. **Miner control failures**
   - Check miner API port (default: 4028)
   - Verify miner firmware supports the required commands
   - Check network connectivity to miners
