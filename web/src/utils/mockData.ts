import {
  HealthResponse,
  StatusResponse,
  MPCDecisionInfo,
  WebSocketMessage,
} from "../types/api";

/**
 * Generates realistic mock data for the Energy Management System
 */

// Helper function to get current hour
function getCurrentHour(): number {
  return new Date().getHours();
}

// Helper function to generate timestamps (returns unix timestamp in seconds)
function getTimestamp(hoursOffset: number = 0): number {
  const now = new Date();
  now.setHours(now.getHours() + hoursOffset);
  now.setMinutes(0);
  now.setSeconds(0);
  now.setMilliseconds(0);
  return Math.floor(now.getTime() / 1000);
}

// Generate realistic solar power based on time of day (0-100 kW)
function generateSolarPower(hour: number): number {
  if (hour < 6 || hour > 20) return 0;
  if (hour < 8) return Math.random() * 15 + 5; // 5-20 kW
  if (hour < 12) return Math.random() * 30 + 60; // 60-90 kW
  if (hour < 16) return Math.random() * 20 + 70; // 70-90 kW
  if (hour < 18) return Math.random() * 30 + 30; // 30-60 kW
  return Math.random() * 10 + 5; // 5-15 kW
}

// Generate realistic battery state
function generateBatteryState(hour: number) {
  const baseSOC = hour < 12 ? 0.3 + hour * 0.03 : 0.9 - (hour - 12) * 0.02;
  const soc = Math.max(0.1, Math.min(0.95, baseSOC + Math.random() * 0.1 - 0.05));

  // Charging during day, discharging during evening (0-20 kW range)
  let power = 0;
  if (hour >= 9 && hour < 15) {
    power = Math.random() * 10 + 5; // Charging: 5-15 kW
  } else if (hour >= 18 && hour < 23) {
    power = -(Math.random() * 15 + 5); // Discharging: -5 to -20 kW
  } else {
    power = Math.random() * 4 - 2; // Idle/small fluctuations: -2 to 2 kW
  }

  return { soc, power };
}

// Generate realistic grid power (0-100 kW range)
function generateGridPower(hour: number): number {
  if (hour >= 10 && hour < 16) {
    // Exporting to grid during peak solar
    return -(Math.random() * 50 + 20); // -20 to -70 kW
  } else if (hour >= 18 && hour < 22) {
    // Importing during evening peak
    return Math.random() * 40 + 40; // 40-80 kW
  }
  return Math.random() * 20 - 10; // -10 to 10 kW
}

// Generate realistic plant active power (consumption) (0-10 kW range)
function generatePlantPower(hour: number): number {
  if (hour >= 8 && hour < 18) {
    // Higher consumption during work hours
    return Math.random() * 3 + 6; // 6-9 kW
  } else if (hour >= 18 && hour < 23) {
    // Evening peak
    return Math.random() * 2 + 7; // 7-9 kW
  }
  // Night time lower consumption
  return Math.random() * 2 + 2; // 2-4 kW
}

// Generate EV charger state (0-30 kW range)
function generateEVChargerState(hour: number) {
  const isCharging = hour >= 10 && hour < 16 && Math.random() > 0.3;
  const vehicleSOC = isCharging
    ? Math.min(0.95, 0.5 + (hour - 10) * 0.07)
    : Math.random() * 0.3 + 0.2;
  const power = isCharging ? Math.random() * 15 + 10 : 0; // 10-25 kW when charging

  return { power, soc: vehicleSOC };
}

// Generate electricity prices (€/kWh) - cached per hour to avoid frequent changes
function generatePrice(hour: number): number {
  // Return cached price if available for this hour
  if (cachedPriceByHour[hour] !== undefined) {
    return cachedPriceByHour[hour];
  }

  // Generate price based on time of day (0.10 to 1 EUR per kWh)
  let price: number;
  if (hour >= 7 && hour < 9) {
    price = Math.random() * 0.30 + 0.60; // Morning peak: 0.60-0.90 €/kWh
  } else if (hour >= 18 && hour < 21) {
    price = Math.random() * 0.40 + 0.60; // Evening peak: 0.60-1.00 €/kWh
  } else if (hour >= 2 && hour < 6) {
    price = Math.random() * 0.20 + 0.10; // Night low: 0.10-0.30 €/kWh
  } else {
    price = Math.random() * 0.30 + 0.30; // Normal: 0.30-0.60 €/kWh
  }

  // Cache the generated price for this hour
  cachedPriceByHour[hour] = price;
  return price;
}

// Generate weather symbol codes
function getWeatherSymbol(hour: number): string {
  const symbols = [
    "clearsky_day",
    "clearsky_night",
    "fair_day",
    "fair_night",
    "partlycloudy_day",
    "partlycloudy_night",
    "cloudy",
    "lightrain",
  ];

  if (hour < 6 || hour > 20) {
    return Math.random() > 0.5 ? "clearsky_night" : "fair_night";
  }
  return symbols[Math.floor(Math.random() * 5)];
}

// Cache for MPC decisions (regenerate every hour)
let cachedMPCDecisions: MPCDecisionInfo[] | null = null;
let mpcCacheTimestamp: number = 0;
const MPC_CACHE_DURATION = 60 * 60 * 1000; // 1 hour in milliseconds

// Cache for miners data (regenerate every hour)
let cachedMiners: Array<{ ip: string; status: string }> | null = null;
let minersCacheTimestamp: number = 0;
const MINERS_CACHE_DURATION = 60 * 60 * 1000; // 1 hour in milliseconds

// Cache for electricity prices (per hour)
const cachedPriceByHour: { [hour: number]: number } = {};

// Generate MPC decisions for the next 24 hours
function generateMPCDecisions(): MPCDecisionInfo[] {
  // Return cached decisions if still valid
  const now = Date.now();
  if (cachedMPCDecisions && (now - mpcCacheTimestamp) < MPC_CACHE_DURATION) {
    return cachedMPCDecisions;
  }

  // Generate fresh decisions
  const decisions: MPCDecisionInfo[] = [];
  const currentHour = getCurrentHour();

  // Start with current realistic battery SOC (as decimal 0-1)
  let currentSOC = Math.max(0.2, Math.min(0.8, 0.4 + Math.random() * 0.2));

  for (let i = 0; i < 24; i++) {
    const hour = (currentHour + i) % 24;
    const timestamp = getTimestamp(i);
    const solarForecast = generateSolarPower(hour); // Already in kW
    const importPrice = generatePrice(hour);
    const exportPrice = importPrice * 0.7; // Export typically lower than import
    const loadForecast = generatePlantPower(hour); // Already in kW
    const cloudCoverage = Math.random() * 100;

    // Battery strategy based on prices and solar
    let batteryCharge = 0;
    let batteryDischarge = 0;

    if (solarForecast > 50 && importPrice < 0.6 && currentSOC < 0.9) {
      // Charge battery during cheap hours with solar
      batteryCharge = Math.random() * 2 + 1;
      currentSOC = Math.min(0.95, currentSOC + batteryCharge * 0.02);
    } else if (importPrice > 0.8 && currentSOC > 0.2) {
      // Discharge during expensive hours
      batteryDischarge = Math.random() * 2 + 0.5;
      currentSOC = Math.max(0.1, currentSOC - batteryDischarge * 0.02);
    } else {
      // Small fluctuations for idle state
      currentSOC = Math.max(0.1, Math.min(0.95, currentSOC + Math.random() * 0.02 - 0.01));
    }

    // Grid strategy - only import OR export, never both
    let gridImport = 0;
    let gridExport = 0;
    
    // Determine net power flow: positive = surplus (export), negative = deficit (import)
    const netPower = solarForecast - loadForecast + batteryDischarge - batteryCharge;
    
    if (netPower > 0 && exportPrice > 0.05) {
      // Surplus power and reasonable export price - export to grid
      gridExport = Math.min(netPower, Math.random() * 40 + 20); // 20-60 kW max
    } else if (netPower < 0) {
      // Power deficit - import from grid
      gridImport = Math.abs(netPower) + Math.random() * 10; // Cover deficit + some margin
    }
    
    const profit = (gridExport * exportPrice - gridImport * importPrice) / 1000;

    decisions.push({
      hour,
      timestamp,
      battery_charge: batteryCharge,
      battery_discharge: batteryDischarge,
      grid_import: gridImport,
      grid_export: gridExport,
      battery_soc: Math.max(0, Math.min(1, currentSOC)),
      profit,
      import_price: importPrice,
      export_price: exportPrice,
      solar_forecast: solarForecast, // Already in kW
      load_forecast: loadForecast,
      cloud_coverage: cloudCoverage,
      weather_symbol: getWeatherSymbol(hour),
    });
  }

  // Cache the new decisions
  cachedMPCDecisions = decisions;
  mpcCacheTimestamp = now;

  return decisions;
}

// Generate mock miner data
function generateMiners() {
  // Return cached miners if still valid
  const now = Date.now();
  if (cachedMiners && (now - minersCacheTimestamp) < MINERS_CACHE_DURATION) {
    return cachedMiners;
  }

  // Generate fresh miners data
  const minerCount = Math.floor(Math.random() * 3) + 2; // 2-4 miners
  const miners = [];
  const statuses = ["RUNNING", "STOPPED", "IDLE"];

  for (let i = 0; i < minerCount; i++) {
    miners.push({
      ip: `192.168.1.${100 + i}`,
      status: statuses[Math.floor(Math.random() * statuses.length)],
    });
  }

  // Cache the new miners data
  cachedMiners = miners;
  minersCacheTimestamp = now;

  return miners;
}

// Calculate sunrise and sunset times
function getSunTimes() {
  const now = new Date();
  const sunrise = new Date(now);
  sunrise.setHours(6, 30 + Math.floor(Math.random() * 30), 0, 0);

  const sunset = new Date(now);
  sunset.setHours(18, 30 + Math.floor(Math.random() * 60), 0, 0);

  return {
    sunrise: sunrise.toISOString(),
    sunset: sunset.toISOString(),
  };
}

// Calculate solar angle (simplified)
function getSolarAngle(): number {
  const hour = getCurrentHour();
  if (hour < 6 || hour > 20) return 0;

  // Peak at noon
  const minutesFromNoon = Math.abs((hour - 12) * 60);
  const maxAngle = 65;
  return Math.max(0, maxAngle - (minutesFromNoon / 360) * maxAngle);
}

/**
 * Generate a complete mock WebSocket message
 */
export function generateMockWebSocketMessage(): WebSocketMessage {
  const hour = getCurrentHour();
  const battery = generateBatteryState(hour);
  const evCharger = generateEVChargerState(hour);
  const sunTimes = getSunTimes();
  const miners = generateMiners();
  const currentPrice = generatePrice(hour);
  const priceLimit = 75;

  const health: HealthResponse = {
    status: "healthy",
    timestamp: new Date().toISOString(),
    version: "1.0.0-demo",
    scheduler: {
      is_running: true,
      miners_count: miners.length,
      has_market_data: true,
      price_limit: priceLimit,
      network: "Demo Network",
      mpc_decisions: generateMPCDecisions(),
    },
    system: {
      uptime: "5d 12h 34m",
      goroutines: 42,
    },
    ems: {
      current_pv_power: generateSolarPower(hour),
      ess_power: battery.power,
      ess_soc: battery.soc,
      grid_sensor_status: 1,
      grid_sensor_active_power: generateGridPower(hour),
      plant_active_power: generatePlantPower(hour),
      dc_charger_output_power: evCharger.power,
      dc_charger_vehicle_soc: evCharger.soc,
    },
    sun: {
      solar_angle: getSolarAngle(),
      sunrise: sunTimes.sunrise,
      sunset: sunTimes.sunset,
    },
  };

  const status: StatusResponse = {
    scheduler_status: {
      is_running: true,
      miners_count: miners.length,
      has_market_data: true,
    },
    miners: {
      count: miners.length,
      list: miners,
    },
    price_data: {
      has_document: true,
      current_avg_price: currentPrice,
      current: currentPrice,
      limit: priceLimit,
    },
    timestamp: new Date().toISOString(),
  };

  return {
    type: "status_update",
    health,
    status,
  };
}

/**
 * Create a mock WebSocket-like interface that emits periodic updates
 */
export function createMockWebSocket(
  onMessage: (data: WebSocketMessage) => void,
  interval: number = 10000
): { close: () => void } {
  // Send initial message immediately
  setTimeout(() => {
    onMessage(generateMockWebSocketMessage());
  }, 100);

  // Set up periodic updates
  const intervalId = setInterval(() => {
    onMessage(generateMockWebSocketMessage());
  }, interval);

  return {
    close: () => clearInterval(intervalId),
  };
}
