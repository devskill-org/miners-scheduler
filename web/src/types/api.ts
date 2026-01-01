export interface SchedulerStatus {
  is_running: boolean;
  miners_count: number;
  has_market_data: boolean;
  price_limit: number;
  network: string;
}

export interface HealthResponse {
  status: string;
  timestamp: string;
  version: string;
  scheduler: SchedulerStatus;
  system: {
    uptime: string;
    goroutines: number;
  };
  ems: {
    current_pv_power: number;
    ess_power: number;
    ess_soc: number;
    grid_sensor_status: number;
    grid_sensor_active_power: number;
    plant_active_power: number;
    dc_charger_output_power: number;
    dc_charger_vehicle_soc: number;
  };
  sun: {
    solar_angle: number;
    sunrise: string;
    sunset: string;
  };
}

export interface StatusResponse {
  scheduler_status: {
    is_running: boolean;
    miners_count: number;
    has_market_data: boolean;
  };
  miners: {
    count: number;
    list: Array<{
      ip: string;
      status: string;
    }>;
  };
  price_data: {
    has_document: boolean;
    current_avg_price?: number;
    current?: number;
    limit?: number;
  };
  timestamp: string;
}

export interface WebSocketMessage {
  type: string;
  health: HealthResponse;
  status: StatusResponse;
}
