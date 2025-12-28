# Power Consumption Control

This document describes the power consumption control feature for the miners scheduler.

## Overview

The power consumption control feature allows the scheduler to manage miners based on available PV (photovoltaic) power and configurable power limits. This ensures that the total power consumed by all miners does not exceed the available solar power or a configured maximum power limit.

## Configuration

The following configuration options have been added to control power consumption:

### Power Limits

- **`miners_power_limit`** (float64, kW): Maximum total power limit for all miners. Default: `30.0` kW
  - This is the absolute maximum power that all miners combined can consume
  - Even if more PV power is available, miners will not exceed this limit

### Miner Power Consumption by Mode

Each miner consumes different amounts of power depending on its operational mode:

- **`miner_power_standby`** (float64, kW): Power consumption in standby mode. Default: `0.05` kW (50 W)
- **`miner_power_eco`** (float64, kW): Power consumption in eco mode. Default: `0.8` kW (800 W)
- **`miner_power_standard`** (float64, kW): Power consumption in standard mode. Default: `1.6` kW (1600 W)
- **`miner_power_super`** (float64, kW): Power consumption in super mode. Default: `1.8` kW (1800 W)

These values should be adjusted based on your specific miner hardware specifications.

### Feature Toggle

- **`use_pv_power_control`** (bool): Enable/disable PV power-based control. Default: `false`
  - When `true`: The scheduler will actively manage miners based on available PV power
  - When `false`: The scheduler uses only price-based control (legacy behavior)

## How It Works

### Power-Based Control Logic

When `use_pv_power_control` is enabled, the scheduler implements the following logic:

1. **Effective Power Limit Calculation**
   - The effective limit is the minimum of:
     - Available PV power (read from the plant Modbus interface)
     - Configured `miners_power_limit`
   - Formula: `effectiveLimit = min(availablePVPower, minersPowerLimit)`

2. **Total Power Monitoring**
   - The scheduler continuously calculates the total power consumption of all miners
   - Power consumption is based on each miner's current state (standby/mining) and work mode (eco/standard/super)

3. **Power Limit Enforcement**
   - If total power consumption exceeds the effective limit, the scheduler will:
     - Identify the highest power-consuming miners
     - Progressively decrease their work modes: Super → Standard → Eco → Standby
     - Continue until total power is within the effective limit

### Integration with Existing Controls

The power control system works alongside existing features:

#### Price-Based Control
- Price-based control still operates when power limits allow
- When waking up a miner due to low electricity prices, the scheduler first checks if there's sufficient power budget
- If waking up would exceed power limits, the miner remains in standby

#### Fan Speed Control (Temperature Management)
- The existing fan speed-based work mode adjustment continues to operate
- Before increasing a miner's work mode based on low fan speed, the scheduler checks power limits
- Work mode increases are blocked if they would exceed available power

### Functions Modified

#### `manageMiners()`
This function now:
1. Checks if PV power control is enabled
2. Reads current available PV power
3. Calculates total power consumption
4. If exceeding limits, calls `adjustMinersForPowerLimit()` to reduce consumption
5. When price allows waking up miners, verifies power budget is available
6. Proceeds with standard price-based control if power allows

#### `runStateCheck()`
This function now:
1. First checks power limits before any adjustments
2. If exceeding limits, reduces miner modes to comply
3. Before increasing miner modes based on fan speed, verifies power budget
4. Continues with fan speed-based adjustments if power allows

#### New Helper Functions

- **`getMinerPowerConsumption(state, workMode)`**: Returns power consumption in kW for a given miner state and mode
- **`calculateTotalPowerConsumption(minersList)`**: Calculates total power consumption of all miners in kW
- **`adjustMinersForPowerLimit(minersList, powerLimit)`**: Adjusts miner modes to stay within power limit (power limit in kW)
- **`GetCurrentPVPower()`**: Retrieves current PV power from the plant Modbus interface in kW

## Example Configuration

```json
{
  "use_pv_power_control": true,
  "miners_power_limit": 30.0,
  "miner_power_standby": 0.05,
  "miner_power_eco": 0.8,
  "miner_power_standard": 1.6,
  "miner_power_super": 1.8,
  "plant_modbus_address": "192.168.1.100:502"
}
```

## Behavior Examples

### Example 1: Insufficient PV Power
- Available PV power: 4.0 kW
- Miners power limit: 30.0 kW
- Effective limit: 4.0 kW (minimum of the two)
- Two miners in Standard mode (1.6 kW each) = 3.2 kW total
- Result: Both miners can operate at Standard mode

### Example 2: Exceeding Power Limit
- Available PV power: 10.0 kW
- Miners power limit: 5.0 kW
- Effective limit: 5.0 kW
- Three miners in Standard mode (1.6 kW each) = 4.8 kW total
- Result: All miners can operate at Standard mode within the limit

### Example 3: Cannot Wake Up Due to Power
- Available PV power: 2.0 kW
- Current: Two miners in Eco mode (0.8 kW each) = 1.6 kW total
- Price drops below limit → scheduler wants to wake up a third miner
- Adding another miner in Eco would require 2.4 kW
- Result: Third miner remains in standby due to insufficient power

## Monitoring

The scheduler logs provide detailed information about power control decisions:

```
PV Power Control: Available PV power: 8.00 kW, Miners power limit: 30.00 kW
Current total power consumption: 6.00 kW, Effective limit: 8.00 kW
Power consumption exceeds limit, adjusting miners...
Power limit: setting miner 192.168.1.10:4028 to Eco mode
Power consumption now within limit: 5.50 kW <= 8.00 kW
```

## Best Practices

1. **Set Accurate Power Values**: Measure or obtain from manufacturer specifications the actual power consumption of your miners in each mode
2. **Configure Appropriate Limits**: Set `miners_power_limit` based on your electrical installation capacity
3. **Monitor PV Power**: Ensure `plant_modbus_address` is correctly configured for accurate PV power readings
4. **Test in Dry-Run Mode**: Use `dry_run: true` to verify behavior before enabling actual control
5. **Consider Safety Margin**: Set power limits slightly below maximum capacity to provide safety margin

## Disabling Power Control

To disable power control and revert to price-only based control:

```json
{
  "use_pv_power_control": false
}
```

The scheduler will then only use price-based control without considering power consumption limits.