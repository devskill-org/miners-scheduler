# Register Address Definition

## 5.1 Plant Running Information Address Definition (Read-only Register)

The registers below can only be accessed by slave address 247, namely "plant address". To obtain power plant data, inquiries should be sent to address 247.

| No. | Name | Address | QTY | Perm. | Data Type | Gain | Unit | Hybrid Inv. | PV Inv. | Comment |
|-----|------|---------|-----|-------|-----------|------|------|-------------|---------|---------|
| 1 | System time | 30000 | 2 | RO | U32 | 1 | s | √ | √ | Epoch seconds |
| 2 | System time zone | 30002 | 1 | RO | S16 | 1 | min | √ | √ | |
| 3 | EMS work mode | 30003 | 1 | RO | U16 | N/A | N/A | √ | √ | 0: Max self consumption; 1: AI Mode; 2: TOU; 7: Remote EMS mode |
| 4 | [Grid Sensor] Status | 30004 | 1 | RO | U16 | N/A | N/A | √ | √ | (gateway or meter connection status) 0: not connected; 1: connected |
| 5 | [Grid sensor] Active power | 30005 | 2 | RO | S32 | 1000 | kW | √ | √ | Data collected from grid sensor at grid to system checkpoint; >0 buy from grid; <0 sell to grid |
| 6 | [Grid sensor] reactive power | 30007 | 2 | RO | S32 | 1000 | kVar | √ | √ | Data collected from grid sensor at grid to system checkpoint |
| 7 | On/Off Grid status | 30009 | 1 | RO | U16 | N/A | N/A | √ | | 0: on grid; 1: off grid (auto); 2: off grid (manual) |
| 8 | Max active power | 30010 | 2 | RO | U32 | 1000 | kW | √ | √ | This is should be the base value of all active power adjustment actions |
| 9 | Max apparent power | 30012 | 2 | RO | U32 | 1000 | kVar | √ | √ | This is should be the base value of all reactive power adjustment actions |
| 10 | [ESS] SOC | 30014 | 1 | RO | U16 | 10 | % | √ | | |
| 11 | Plant phase A active power | 30015 | 2 | RO | S32 | 1000 | kW | √ | √ | |
| 12 | Plant phase B active power | 30017 | 2 | RO | S32 | 1000 | kW | √ | √ | |
| 13 | Plant phase C active power | 30019 | 2 | RO | S32 | 1000 | kW | √ | √ | |
| 14 | Plant phase A reactive power | 30021 | 2 | RO | S32 | 1000 | kVar | √ | √ | |
| 15 | Plant phase B reactive power | 30023 | 2 | RO | S32 | 1000 | kVar | √ | √ | |
| 16 | Plant phase C reactive power | 30025 | 2 | RO | S32 | 1000 | kVar | √ | √ | |
| 17 | General Alarm1 | 30027 | 1 | RO | U16 | N/A | N/A | √ | √ | If any hybrid inverter has alarm, then this alarm will be set accordingly. Refer to Appendix 2 |
| 18 | General Alarm2 | 30028 | 1 | RO | U16 | N/A | N/A | √ | √ | If any hybrid inverter has alarm, then this alarm will be set accordingly. Refer to Appendix 3 |
| 19 | General Alarm3 | 30029 | 1 | RO | U16 | N/A | N/A | √ | | If any hybrid inverter has alarm, then this alarm will be set accordingly. Refer to Appendix 4 |
| 20 | General Alarm4 | 30030 | 1 | RO | U16 | N/A | N/A | √ | √ | If any hybrid inverter has alarm, then this alarm will be set accordingly. Refer to Appendix 5 |
| 21 | Plant active power | 30031 | 2 | RO | S32 | 1000 | kW | √ | √ | |
| 22 | Plant reactive power | 30033 | 2 | RO | S32 | 1000 | kVar | √ | √ | |
| 23 | Photovoltaic power | 30035 | 2 | RO | S32 | 1000 | kW | √ | √ | |
| 24 | [ESS] power | 30037 | 2 | RO | S32 | 1000 | kW | √ | | <0: discharging; >0: charging |
| 25 | Available max active power | 30039 | 2 | RO | U32 | 1000 | kW | √ | √ | Feed to the ac terminal. Count only the running inverters |
| 26 | Available min active power | 30041 | 2 | RO | U32 | 1000 | kW | √ | | Absorb from the ac terminal. Count only the running inverters |
| 27 | Available max reactive power | 30043 | 2 | RO | U32 | 1000 | kVar | √ | √ | Feed to the ac terminal. Count only the running inverters |
| 28 | Available min reactive power | 30045 | 2 | RO | U32 | 1000 | kVar | √ | √ | Absorb from the ac terminal. Count only the running inverters |
| 29 | [ESS]Available max charging power | 30047 | 2 | RO | U32 | 1000 | kW | √ | | Count only the running inverters |
| 30 | [ESS]Available max discharging power | 30049 | 2 | RO | U32 | 1000 | kW | √ | | Count only the running inverters |
| 31 | Plant running state | 30051 | 1 | RO | U16 | N/A | N/A | √ | √ | Refer to Appendix 1 |
| 32 | [Grid sensor] Phase A active power | 30052 | 2 | RO | S32 | 1000 | kW | √ | √ | Data collected from grid sensor at grid to system checkpoint; >0 buy from grid; <0 sell to grid |
| 33 | [Grid sensor] Phase B active power | 30054 | 2 | RO | S32 | 1000 | kW | √ | √ | Data collected from grid sensor at grid to system checkpoint; >0 buy from grid; <0 sell to grid |
| 34 | [Grid sensor] Phase C active power | 30056 | 2 | RO | S32 | 1000 | kW | √ | √ | Data collected from grid sensor at grid to system checkpoint; >0 buy from grid; <0 sell to grid |
| 35 | [Grid sensor] Phase A reactive power | 30058 | 2 | RO | S32 | 1000 | kVar | √ | √ | Data collected from grid sensor at grid to system checkpoint |
| 36 | [Grid sensor] Phase B reactive power | 30060 | 2 | RO | S32 | 1000 | kVar | √ | √ | Data collected from grid sensor at grid to system checkpoint |
| 37 | [Grid sensor] Phase C reactive power | 30062 | 2 | RO | S32 | 1000 | kVar | √ | √ | Data collected from grid sensor at grid to system checkpoint |
| 38 | [ESS]Available max charging capacity | 30064 | 2 | RO | U32 | 100 | kWh | √ | | Count only the running inverters |
| 39 | [ESS]Available max discharging capacity | 30066 | 2 | RO | U32 | 100 | kWh | √ | | Count only the running inverters |
| 40 | [ESS] Rated charging power | 30068 | 2 | RO | U32 | 1000 | kW | √ | | |
| 41 | [ESS] Rated discharging power | 30070 | 2 | RO | U32 | 1000 | kW | √ | | |
| 42 | General Alarm5 | 30072 | 1 | RO | U16 | N/A | N/A | √ | | If any hybrid inverter has alarm, then this alarm will be set accordingly. Refer to Appendix 11 |
| 43 | Reserved | 30073 | 10 | RO | N/A | N/A | N/A | | | |
| 44 | [ESS] rated energy capacity | 30083 | 2 | RO | U32 | 100 | kWh | √ | | |
| 45 | [ESS] charge Cut-Off SOC | 30085 | 1 | RO | U16 | 10 | % | √ | | |
| 46 | [ESS] discharge Cut-Off SOC | 30086 | 1 | RO | U16 | 10 | % | √ | | |
| 47 | [ESS] SOH | 30087 | 1 | RO | U16 | 10 | % | √ | | This value is the weighted average of the SOH of all ESS devices in the power plant, with each rated capacity as the weight. |

---

## 5.2 Plant Parameter Setting Address Definition (Holding Register)

The registers below can only be accessed by slave address 0 or 247. To modify plant-level registers, send commands to address 0 or 247. When sending commands to address 0, the device will only execute and will not reply. When sending commands to address 247, the device will both execute and respond.

**Note:** Power control related registers not explicitly mentioned in the "Comment" will take effect only when the remote EMS control mode value is 0.

| No. | Name | Address | QTY | Perm. | Data Type | Gain | Unit | Hybrid Inv. | PV Inv. | Comment |
|-----|------|---------|-----|-------|-----------|------|------|-------------|---------|---------|
| 1 | Start/Stop | 40000 | 1 | WO | U16 | N/A | N/A | √ | √ | 0: Stop; 1: Start |
| 2 | Active power fixed adjustment target value | 40001 | 2 | RW | S32 | 1000 | kW | √ | √ | |
| 3 | Reactive power fixed adjustment target value | 40003 | 2 | RW | S32 | 1000 | kVar | √ | √ | Range: [-60.00 * base value, 60.00 * base value]. Takes effect globally regardless of the EMS operating mode. |
| 4 | Active power percentage adjustment target value | 40005 | 1 | RW | S16 | 100 | % | √ | √ | Range: [-100.00, 100.00] |
| 5 | Q/S adjustment target value | 40006 | 1 | RW | S16 | 100 | % | √ | √ | Range: [-60.00, 60.00]. Takes effect globally regardless of the EMS operating mode. |
| 6 | Power factor adjustment target value | 40007 | 1 | RW | S16 | 1000 | N/A | √ | √ | Range: (-1, -0.8] U [0.8, 1]. Grid Sensor needed. Takes effect globally regardless of the EMS operating mode. |
| 7 | Phase A active power fixed adjustment target value | 40008 | 2 | RW | S32 | 1000 | kW | √ | | Valid only when output type is L1/L2/L3/N |
| 8 | Phase B active power fixed adjustment target value | 40010 | 2 | RW | S32 | 1000 | kW | √ | | Valid only when output type is L1/L2/L3/N |
| 9 | Phase C active power fixed adjustment target value | 40012 | 2 | RW | S32 | 1000 | kW | √ | | Valid only when output type is L1/L2/L3/N |
| 10 | Phase A reactive power fixed adjustment target value | 40014 | 2 | RW | S32 | 1000 | kVar | √ | | Valid only when output type is L1/L2/L3/N |
| 11 | Phase B reactive power fixed adjustment target value | 40016 | 2 | RW | S32 | 1000 | kVar | √ | | Valid only when output type is L1/L2/L3/N |
| 12 | Phase C reactive power fixed adjustment target value | 40018 | 2 | RW | S32 | 1000 | kVar | √ | | Valid only when output type is L1/L2/L3/N |
| 13 | Phase A Active power percentage adjustment target value | 40020 | 1 | RW | S16 | 100 | % | √ | | Valid only when output type is L1/L2/L3/N. Range: [-100.00, 100.00] |
| 14 | Phase B Active power percentage adjustment target value | 40021 | 1 | RW | S16 | 100 | % | √ | | Valid only when output type is L1/L2/L3/N. Range: [-100.00, 100.00] |
| 15 | Phase C Active power percentage adjustment target value | 40022 | 1 | RW | S16 | 100 | % | √ | | Valid only when output type is L1/L2/L3/N. Range: [-100.00, 100.00] |
| 16 | Phase A Q/S fixed adjustment target value | 40023 | 1 | RW | S16 | 100 | % | √ | | Valid only when output type is L1/L2/L3/N. Range: [-60.00, 60.00] |
| 17 | Phase B Q/S fixed adjustment target value | 40024 | 1 | RW | S16 | 100 | % | √ | | Valid only when output type is L1/L2/L3/N. Range: [-60.00, 60.00] |
| 18 | Phase C Q/S fixed adjustment target value | 40025 | 1 | RW | S16 | 100 | % | √ | | Valid only when output type is L1/L2/L3/N. Range: [-60.00, 60.00] |
| 19 | Reserved | 40026 | 3 | RW | N/A | N/A | N/A | | | |
| 20 | Remote EMS enable | 40029 | 1 | RW | U16 | N/A | N/A | √ | √ | 0: disabled; 1: enabled. When needed to control EMS remotely, this register needs to be enabled. When enabled, the plant's EMS work mode (30003) will switch to remote EMS. |
| 21 | Independent phase power control enable | 40030 | 1 | RW | U16 | N/A | N/A | √ | | Valid only when output type is L1/L2/L3/N. To enable independent phase control, this parameter must be enabled. 0: disabled; 1: enabled |
| 22 | Remote EMS control mode | 40031 | 1 | RW | U16 | N/A | N/A | √ | √ | Mode values' definition refer to Appendix 6 |
| 23 | ESS max charging limit | 40032 | 2 | RW | U32 | 1000 | kW | √ | | [0, Rated ESS charging power]. Takes effect when Remote EMS control mode (40031) is 3 or 4. |
| 24 | ESS max discharging limit | 40034 | 2 | RW | U32 | 1000 | kW | √ | | [0, Rated ESS discharging power]. Takes effect when Remote EMS control mode (40031) is 5 or 6. |
| 25 | PV max power limit | 40036 | 2 | RW | U32 | 1000 | kW | √ | | Takes effect when Remote EMS control mode (40031) is 3, 4, 5 or 6. |
| 26 | [Grid Point]Maximum export limitation | 40038 | 2 | RW | U32 | 1000 | kW | √ | √ | Grid Sensor needed. Takes effect globally regardless of the EMS operating mode. |
| 27 | [Grid Point] Maximum import limitation | 40040 | 2 | RW | U32 | 1000 | kW | √ | √ | Grid Sensor needed. Takes effect globally regardless of the EMS operating mode. |
| 28 | PCS maximum export limitation | 40042 | 2 | RW | U32 | 1000 | kW | √ | √ | Range:[0, 0xFFFFFFFE]. With value 0xFFFFFFFF, register is not valid. In all other cases, Takes effect globally. |
| 29 | PCS maximum import limitation | 40044 | 2 | RW | U32 | 1000 | kW | √ | √ | Range:[0, 0xFFFFFFFE]. With value 0xFFFFFFFF, register is not valid. In all other cases, Takes effect globally. |

---

## 5.3 Hybrid Inverter Running Information Address Definition (Read-only Register)

The registers below can only be accessed with a valid Hybrid inverter's Modbus slave address (1-246). When using PV string related registers, please refer to the PV count listed in Table 2-1 in Chapter 2, to ensure if the register is available.

| No. | Name | Address | QTY | Perm. | Data Type | Gain | Unit | Hybrid Inv. | PV Inv. | Comment |
|-----|------|---------|-----|-------|-----------|------|------|-------------|---------|---------|
| 1 | Model type | 30500 | 15 | RO | STRING | N/A | N/A | √ | √ | |
| 2 | Serial number | 30515 | 10 | RO | STRING | N/A | N/A | √ | √ | |
| 3 | Machine firmware version | 30525 | 15 | RO | STRING | N/A | N/A | √ | √ | |
| 4 | Rated active power | 30540 | 2 | RO | U32 | 1000 | kW | √ | √ | |
| 5 | Max. apparent power | 30542 | 2 | RO | U32 | 1000 | kVA | √ | √ | |
| 6 | Max. active power | 30544 | 2 | RO | U32 | 1000 | kW | √ | √ | |
| 7 | Max. absorption power | 30546 | 2 | RO | U32 | 1000 | kW | √ | | |
| 8 | Rated battery capacity | 30548 | 2 | RO | U32 | 100 | kWh | √ | | |
| 9 | [ESS]Rated charge power | 30550 | 2 | RO | U32 | 1000 | kW | √ | | |
| 10 | [ESS]Rated discharge power | 30552 | 2 | RO | U32 | 1000 | kW | √ | | |
| 11 | Reserved | 30554 | 12 | RO | N/A | N/A | N/A | | | |
| 12 | [ESS]Daily charge energy | 30566 | 2 | RO | U32 | 100 | kWh | √ | | |
| 13 | [ESS]Accumulated charge energy | 30568 | 4 | RO | U64 | 100 | kWh | √ | | |
| 14 | [ESS]Daily discharge energy | 30572 | 2 | RO | U32 | 100 | kWh | √ | | |
| 15 | [ESS]Accumulated discharge energy | 30574 | 4 | RO | U64 | 100 | kWh | √ | | |
| 16 | Running state | 30578 | 1 | RO | U16 | N/A | N/A | √ | √ | Refer to Appendix 1 |
| 17 | Max.active power adjustment value | 30579 | 2 | RO | S32 | 1000 | kW | √ | √ | |
| 18 | Min. active power adjustment value | 30581 | 2 | RO | S32 | 1000 | kW | √ | | |
| 19 | Max. reactive power adjustment value fed to the ac terminal | 30583 | 2 | RO | U32 | 1000 | kVar | √ | √ | |
| 20 | Max. reactive power adjustment value absorbed from the ac terminal | 30585 | 2 | RO | U32 | 1000 | kVar | √ | √ | |
| 21 | Active power | 30587 | 2 | RO | S32 | 1000 | kW | √ | √ | |
| 22 | Reactive power | 30589 | 2 | RO | S32 | 1000 | kVar | √ | √ | |
| 23 | [ESS]Max. battery charge power | 30591 | 2 | RO | U32 | 1000 | kW | √ | | |
| 24 | [ESS]Max. battery discharge power | 30593 | 2 | RO | U32 | 1000 | kW | √ | | |
| 25 | [ESS]Available battery charge Energy | 30595 | 2 | RO | U32 | 100 | kWh | √ | | |
| 26 | [ESS]Available battery discharge Energy | 30597 | 2 | RO | U32 | 100 | kWh | √ | | |
| 27 | [ESS] Charge / discharge power | 30599 | 2 | RO | S32 | 1000 | kW | √ | | |
| 28 | [ESS]Battery SOC | 30601 | 1 | RO | U16 | 10 | % | √ | | |
| 29 | [ESS]Battery SOH | 30602 | 1 | RO | U16 | 10 | % | √ | | |
| 30 | [ESS]Average cell temperature | 30603 | 1 | RO | S16 | 10 | ℃ | √ | | |
| 31 | [ESS] Average cell voltage | 30604 | 1 | RO | U16 | 1000 | V | √ | | |
| 32 | Alarm1 | 30605 | 1 | RO | U16 | N/A | N/A | √ | √ | Refer to Appendix 2 |
| 33 | Alarm2 | 30606 | 1 | RO | U16 | N/A | N/A | √ | √ | Refer to Appendix 3 |
| 34 | Alarm3 | 30607 | 1 | RO | U16 | N/A | N/A | √ | | Refer to Appendix 4 |
| 35 | Alarm4 | 30608 | 1 | RO | U16 | N/A | N/A | √ | √ | Refer to Appendix 5 |
| 36 | Alarm5 | 30609 | 1 | RO | U16 | N/A | N/A | √ | | Refer to Appendix 11 |
| 37 | Reserved | 30610 | 10 | RO | N/A | N/A | N/A | | | |
| 38 | [ESS]Maximum battery (cluster) temperature | 30620 | 1 | RO | S16 | 10 | ℃ | √ | | |
| 39 | [ESS]Minimum battery (cluster) temperature | 30621 | 1 | RO | S16 | 10 | ℃ | √ | | |
| 40 | [ESS] Maximum battery (cluster) cell voltage | 30622 | 1 | RO | U16 | 1000 | V | √ | | |
| 41 | [ESS] Minimum battery (cluster) cell voltage | 30623 | 1 | RO | U16 | 1000 | V | √ | | |
| 42 | Rated grid voltage | 31000 | 1 | RO | U16 | 10 | V | √ | √ | |
| 43 | Rated grid frequency | 31001 | 1 | RO | U16 | 100 | Hz | √ | √ | |
| 44 | Grid frequency | 31002 | 1 | RO | U16 | 100 | Hz | √ | √ | |
| 45 | [PCS] Internal temperature | 31003 | 1 | RO | S16 | 10 | ℃ | √ | √ | |
| 46 | Output type | 31004 | 1 | RO | U16 | N/A | N/A | √ | √ | 0: L/N; 1: L1/L2/L3; 2: L1/L2/L3/N; 3: L1/L2/N |
| 47 | A-B line voltage | 31005 | 2 | RO | U32 | 100 | V | √ | √ | Invalid when output type is L/N, L1/L2/N, or L1/L2/N |
| 48 | B-C line voltage | 31007 | 2 | RO | U32 | 100 | V | √ | √ | |
| 49 | C-A line voltage | 31009 | 2 | RO | U32 | 100 | V | √ | √ | |
| 50 | Phase A voltage | 31011 | 2 | RO | U32 | 100 | V | √ | √ | When output type is L/N, refers to "Phase voltage" |
| 51 | Phase B voltage | 31013 | 2 | RO | U32 | 100 | V | √ | √ | Invalid when output type is L/N, L1/L2/N, or L1/L2/N |
| 52 | Phase C voltage | 31015 | 2 | RO | U32 | 100 | V | √ | √ | |
| 53 | Phase A current | 31017 | 2 | RO | S32 | 100 | A | √ | √ | When output type is L/N, refers to "Phase current" |
| 54 | Phase B current | 31019 | 2 | RO | S32 | 100 | A | √ | √ | Invalid when output type is L/N, L1/L2/N, or L1/L2/N |
| 55 | Phase C current | 31021 | 2 | RO | S32 | 100 | A | √ | √ | |
| 56 | Power factor | 31023 | 1 | RO | U16 | 1000 | N/A | √ | √ | |
| 57 | PACK count | 31024 | 1 | RO | U16 | 1 | N/A | √ | | |
| 58 | PV string count | 31025 | 1 | RO | U16 | 1 | N/A | √ | √ | |
| 59 | MPPT count | 31026 | 1 | RO | U16 | 1 | N/A | √ | √ | |
| 60 | PV1 voltage | 31027 | 1 | RO | S16 | 10 | V | √ | √ | Please refer to the PV count listed in Table 2-1 in chapter 2, to ensure if the register is available. |
| 61 | PV1 current | 31028 | 1 | RO | S16 | 100 | A | √ | √ | |
| 62 | PV2 voltage | 31029 | 1 | RO | S16 | 10 | V | √ | √ | |
| 63 | PV2 current | 31030 | 1 | RO | S16 | 100 | A | √ | √ | |
| 64 | PV3 voltage | 31031 | 1 | RO | S16 | 10 | V | √ | √ | |
| 65 | PV3 current | 31032 | 1 | RO | S16 | 100 | A | √ | √ | |
| 66 | PV4 voltage | 31033 | 1 | RO | S16 | 10 | V | √ | √ | |
| 67 | PV4 current | 31034 | 1 | RO | S16 | 100 | A | √ | √ | |
| 68 | PV power | 31035 | 2 | RO | S32 | 1000 | kW | √ | √ | |
| 69 | Insulation resistance | 31037 | 1 | RO | U16 | 1000 | MΩ | √ | √ | |
| 70 | Startup time | 31038 | 2 | RO | U32 | 1 | s | √ | √ | |
| 71 | Shutdown time | 31040 | 2 | RO | U32 | 1 | s | √ | √ | |
| 72 | PV5 voltage | 31042 | 1 | RO | S16 | 10 | V | √ | √ | Please refer to the PV count listed in Table 2-1 in chapter 2, to ensure if the register is available. |
| 73 | PV5 current | 31043 | 1 | RO | S16 | 100 | A | √ | √ | |
| 74 | PV6 voltage | 31044 | 1 | RO | S16 | 10 | V | √ | √ | |
| 75 | PV6 current | 31045 | 1 | RO | S16 | 100 | A | √ | √ | |
| 76 | PV7 voltage | 31046 | 1 | RO | S16 | 10 | V | √ | √ | |
| 77 | PV7 current | 31047 | 1 | RO | S16 | 100 | A | √ | √ | |
| 78 | PV8 voltage | 31048 | 1 | RO | S16 | 10 | V | √ | √ | |
| 79 | PV8 current | 31049 | 1 | RO | S16 | 100 | A | √ | √ | |
| 80 | PV9 voltage | 31050 | 1 | RO | S16 | 10 | V | √ | √ | |
| 81 | PV9 current | 31051 | 1 | RO | S16 | 100 | A | √ | √ | |
| 82 | PV10 voltage | 31052 | 1 | RO | S16 | 10 | V | √ | √ | |
| 83 | PV10 current | 31053 | 1 | RO | S16 | 100 | A | √ | √ | |
| 84 | PV11 voltage | 31054 | 1 | RO | S16 | 10 | V | √ | √ | |
| 85 | PV11 current | 31055 | 1 | RO | S16 | 100 | A | √ | √ | |
| 86 | PV12 voltage | 31056 | 1 | RO | S16 | 10 | V | √ | √ | |
| 87 | PV12 current | 31057 | 1 | RO | S16 | 100 | A | √ | √ | |
| 88 | PV13 voltage | 31058 | 1 | RO | S16 | 10 | V | √ | √ | |
| 89 | PV13 current | 31059 | 1 | RO | S16 | 100 | A | √ | √ | |
| 90 | PV14 voltage | 31060 | 1 | RO | S16 | 10 | V | √ | √ | |
| 91 | PV14 current | 31061 | 1 | RO | S16 | 100 | A | √ | √ | |
| 92 | PV15 voltage | 31062 | 1 | RO | S16 | 10 | V | √ | √ | |
| 93 | PV15 current | 31063 | 1 | RO | S16 | 100 | A | √ | √ | |
| 94 | PV16 voltage | 31064 | 1 | RO | S16 | 10 | V | √ | √ | |
| 95 | PV16 current | 31065 | 1 | RO | S16 | 100 | A | √ | √ | |
| 96 | [DC Charger] Vehicle battery voltage | 31500 | 1 | RO | U16 | 10 | V | √ | | |
| 97 | [DC Charger] Charging current | 31501 | 1 | RO | U16 | 10 | A | √ | | |
| 98 | [DC Charger] Output power | 31502 | 2 | RO | S32 | 1000 | kW | √ | | |
| 99 | [DC Charger] Vehicle SOC | 31504 | 1 | RO | U16 | 10 | % | √ | | |
| 100 | [DC Charger] Current charging capacity | 31505 | 2 | RO | U32 | 100 | kWh | √ | | Single time |
| 101 | [DC Charger] Current charging duration | 31507 | 2 | RO | U32 | 1 | s | √ | | Single time |

---

## 5.4 Hybrid Inverter Parameter Setting Address Definition (Holding Register)

The registers below can only be accessed with a valid Hybrid inverter's Modbus slave address (1-246).

| No. | Name | Address | QTY | Perm. | Data Type | Gain | Unit | Hybrid Inv. | PV Inv. | Comment |
|-----|------|---------|-----|-------|-----------|------|------|-------------|---------|---------|
| 1 | Start/Stop | 40500 | 1 | WO | U16 | N/A | N/A | √ | √ | 0: Stop; 1: Start |
| 2 | Grid code | 40501 | 1 | RW | U16 | N/A | N/A | √ | √ | |
| 3 | [DC Charger] Start/Stop | 41000 | 1 | WO | U16 | N/A | N/A | √ | | 0: Start; 1: Stop |
| 4 | Remote EMS dispatch enable | 41500 | 1 | RW | U16 | N/A | N/A | √ | | 0: disabled; 1: enabled. The enabled inverter only reacts on power control command from register: 41501, 41503, 41505, 41506, 40507. |
| 5 | Active power fixed value adjustment | 41501 | 2 | RW | S32 | 1000 | kW | √ | | |
| 6 | Reactive power fixed value adjustment | 41503 | 2 | RW | S32 | 1000 | kVar | √ | | |
| 7 | Active power percentage adjustment | 41505 | 1 | RW | S16 | 100 | % | √ | | |
| 8 | Reactive power Q/S adjustment | 41506 | 1 | RW | S16 | 100 | % | √ | | |
| 9 | Power factor adjustment | 41507 | 1 | RW | S16 | 1000 | N/A | √ | | |

---

## 5.5 AC-Charger Running Information Address Definition (Read-only Register)

The registers below can only be accessed with a valid AC-Charger's Modbus slave address (1-246). And are only applicable for "EVAC" devices.

| No. | Name | Address | QTY | Perm. | Data Type | Gain | Unit | Comment |
|-----|------|---------|-----|-------|-----------|------|------|---------|
| 1 | System state | 32000 | 1 | RO | U16 | N/A | N/A | System states according to IEC61851-1 definition. Refer to Appendix 7. |
| 2 | Total energy consumed | 32001 | 2 | RO | U32 | 100 | kWh | |
| 3 | Charging power | 32003 | 2 | RO | S32 | 1000 | kW | |
| 4 | Rated power | 32005 | 2 | RO | U32 | 1000 | kW | |
| 5 | Rated current | 32007 | 2 | RO | S32 | 100 | A | |
| 6 | Rated voltage | 32009 | 1 | RO | U16 | 10 | V | |
| 7 | AC-Charger input breaker rated current | 32010 | 2 | RO | S32 | 100 | A | |
| 8 | Alarm1 | 32012 | 1 | RO | U16 | N/A | N/A | Refer to Appendix 8 |
| 9 | Alarm2 | 32013 | 1 | RO | U16 | N/A | N/A | Refer to Appendix 9 |
| 10 | Alarm3 | 32014 | 1 | RO | U16 | N/A | N/A | Refer to Appendix 10 |

---

## 5.6 AC-Charger Parameter Setting Address Definition (Holding Register)

The registers below can only be accessed with a valid AC-Charger's Modbus slave address (1-246). And are only applicable for "EVAC" devices.

| No. | Name | Address | QTY | Perm. | Data Type | Gain | Unit | Comment |
|-----|------|---------|-----|-------|-----------|------|------|---------|
| 1 | Start/Stop | 42000 | 1 | WO | U16 | N/A | N/A | 0: Start; 1: Stop |
| 2 | Charger output current | 42001 | 2 | RW | U32 | 100 | N/A | [6, X]. X is the smaller value between the rated current and the AC-Charger input breaker rated current. |