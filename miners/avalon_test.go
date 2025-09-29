package miners

import (
	"encoding/json"
	"os"
	"testing"
)

func TestAvalonQLiteStatParsing(t *testing.T) {
	// Read the test data file
	data, err := os.ReadFile("../test_data/avalon_litestat.json")
	if err != nil {
		t.Fatalf("Failed to read test data file: %v", err)
	}

	// Parse the JSON
	var liteStat AvalonQLiteStats
	if err := json.Unmarshal(data, &liteStat); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Test STATUS parsing
	if len(liteStat.Status) != 1 {
		t.Errorf("Expected 1 status item, got %d", len(liteStat.Status))
	}

	status := liteStat.Status[0]
	if status.Status != "S" {
		t.Errorf("Expected status 'S', got '%s'", status.Status)
	}
	if status.When != 1757157198 {
		t.Errorf("Expected when 1757157198, got %d", status.When)
	}
	if status.Code != 70 {
		t.Errorf("Expected code 70, got %d", status.Code)
	}
	if status.Msg != "CGMiner stats" {
		t.Errorf("Expected msg 'CGMiner stats', got '%s'", status.Msg)
	}
	if status.Description != "cgminer 4.11.1" {
		t.Errorf("Expected description 'cgminer 4.11.1', got '%s'", status.Description)
	}

	// Test STATS parsing
	if len(liteStat.Stats) != 1 {
		t.Errorf("Expected 1 stats item, got %d", len(liteStat.Stats))
	}

	stats := liteStat.Stats[0].MMIDSummary
	if stats == nil {
		t.Fatalf("ParsedStats is nil")
	}

	// Test version fields
	if stats.Ver != "Q-25052801_14a19a2" {
		t.Errorf("Expected Ver 'Q-25052801_14a19a2', got '%s'", stats.Ver)
	}
	if stats.LVer != "25052801_14a19a2" {
		t.Errorf("Expected LVer '25052801_14a19a2', got '%s'", stats.LVer)
	}
	if stats.BVer != "25052801_14a19a2" {
		t.Errorf("Expected BVer '25052801_14a19a2', got '%s'", stats.BVer)
	}

	// Test hardware info
	if stats.HashMcu0Ver != "Q_hb_v1.1" {
		t.Errorf("Expected HashMcu0Ver 'Q_hb_v1.1', got '%s'", stats.HashMcu0Ver)
	}
	if stats.FanMcuVer != "Q_fb_v1.2" {
		t.Errorf("Expected FanMcuVer 'Q_fb_v1.2', got '%s'", stats.FanMcuVer)
	}
	if stats.CPU != "K230" {
		t.Errorf("Expected CPU 'K230', got '%s'", stats.CPU)
	}
	if stats.FW != "Release" {
		t.Errorf("Expected FW 'Release', got '%s'", stats.FW)
	}
	if stats.DNA != "020100001a0f0bc8" {
		t.Errorf("Expected DNA '020100001a0f0bc8', got '%s'", stats.DNA)
	}

	// Test system state
	if stats.State != AvalonStateMining {
		t.Errorf("Expected State %d (Mining), got %d", AvalonStateMining, stats.State)
	}
	if stats.MemFree != 67660 {
		t.Errorf("Expected MemFree 67660, got %d", stats.MemFree)
	}

	// Test NetFail array
	expectedNetFail := []int64{1757135953, 1757135954, 1757137163, 1757137168, 1757138369, 1757138374, 1757134806, 1757134811}
	if len(stats.NetFail) != len(expectedNetFail) {
		t.Errorf("Expected %d NetFail entries, got %d", len(expectedNetFail), len(stats.NetFail))
	}
	for i, expected := range expectedNetFail {
		if i < len(stats.NetFail) && stats.NetFail[i] != expected {
			t.Errorf("Expected NetFail[%d] %d, got %d", i, expected, stats.NetFail[i])
		}
	}

	// Test network info
	if stats.SSID != "" {
		t.Errorf("Expected empty SSID, got '%s'", stats.SSID)
	}
	if stats.RSSI != 0 {
		t.Errorf("Expected RSSI 0, got %d", stats.RSSI)
	}
	if stats.NetDevType != 0 {
		t.Errorf("Expected NetDevType 0, got %d", stats.NetDevType)
	}

	// Test system status
	if stats.SystemStatus != "Work: In Work, Hash Board: 1" {
		t.Errorf("Expected SystemStatus 'Work: In Work, Hash Board: 1', got '%s'", stats.SystemStatus)
	}

	// Test elapsed time
	if stats.Elapsed != 769980 {
		t.Errorf("Expected Elapsed 769980, got %d", stats.Elapsed)
	}

	// Test boot info
	if stats.BootBy != "0x01.00000000" {
		t.Errorf("Expected BootBy '0x01.00000000', got '%s'", stats.BootBy)
	}

	// Test mining stats
	if stats.LW != 9594440 {
		t.Errorf("Expected LW 9594440, got %d", stats.LW)
	}
	if stats.MH != 1 {
		t.Errorf("Expected MH 1, got %d", stats.MH)
	}
	if stats.DHW != 0 {
		t.Errorf("Expected DHW 0, got %d", stats.DHW)
	}
	if stats.HW != 1 {
		t.Errorf("Expected HW 1, got %d", stats.HW)
	}
	if stats.DH != "3.488%" {
		t.Errorf("Expected DH '3.488%%', got '%s'", stats.DH)
	}

	// Test temperatures
	if stats.ITemp != 49 {
		t.Errorf("Expected ITemp 49, got %d", stats.ITemp)
	}
	if stats.HBITemp != 60 {
		t.Errorf("Expected HBITemp 60, got %d", stats.HBITemp)
	}
	if stats.HBOTemp != 63 {
		t.Errorf("Expected HBOTemp 63, got %d", stats.HBOTemp)
	}
	if stats.TMax != 72 {
		t.Errorf("Expected TMax 72, got %d", stats.TMax)
	}
	if stats.TAvg != 65 {
		t.Errorf("Expected TAvg 65, got %d", stats.TAvg)
	}
	if stats.TarT != 65 {
		t.Errorf("Expected TarT 65, got %d", stats.TarT)
	}

	// Test fan speeds
	if stats.Fan1 != 2129 {
		t.Errorf("Expected Fan1 2129, got %d", stats.Fan1)
	}
	if stats.Fan2 != 2128 {
		t.Errorf("Expected Fan2 2128, got %d", stats.Fan2)
	}
	if stats.Fan3 != 2107 {
		t.Errorf("Expected Fan3 2107, got %d", stats.Fan3)
	}
	if stats.Fan4 != 2100 {
		t.Errorf("Expected Fan4 2100, got %d", stats.Fan4)
	}
	if stats.FanR != 71 {
		t.Errorf("Expected FanR 71, got %d", stats.FanR)
	}

	// Test timing
	if stats.SoftOffTime != 1757170800 {
		t.Errorf("Expected SoftOffTime 1757170800, got %d", stats.SoftOffTime)
	}
	if stats.SoftOnTime != 1757139337 {
		t.Errorf("Expected SoftOnTime 1757139337, got %d", stats.SoftOnTime)
	}

	// Test misc fields
	if stats.Filter != 71618 {
		t.Errorf("Expected Filter 71618, got %d", stats.Filter)
	}
	if stats.FanErr != 0 {
		t.Errorf("Expected FanErr 0, got %d", stats.FanErr)
	}
	if stats.SoloAllowed != 0 {
		t.Errorf("Expected SoloAllowed 0, got %d", stats.SoloAllowed)
	}

	// Test PS array
	expectedPS := []int{0, 1212, 2285, 35, 801, 2286, 822}
	if len(stats.PS) != len(expectedPS) {
		t.Errorf("Expected %d PS entries, got %d", len(expectedPS), len(stats.PS))
	}
	for i, expected := range expectedPS {
		if i < len(stats.PS) && stats.PS[i] != expected {
			t.Errorf("Expected PS[%d] %d, got %d", i, expected, stats.PS[i])
		}
	}

	// Test communication error
	if stats.PCOMM_E != 0 {
		t.Errorf("Expected PCOMM_E 0, got %d", stats.PCOMM_E)
	}

	// Test hash rate fields
	if stats.GHSspd != 50985.10 {
		t.Errorf("Expected GHSspd 50985.10, got %f", stats.GHSspd)
	}
	if stats.DHspd != "3.488%" {
		t.Errorf("Expected DHspd '3.488%%', got '%s'", stats.DHspd)
	}
	if stats.GHSmm != 52838.21 {
		t.Errorf("Expected GHSmm 52838.21, got %f", stats.GHSmm)
	}
	if stats.GHSavg != 9080.65 {
		t.Errorf("Expected GHSavg 9080.65, got %f", stats.GHSavg)
	}
	if stats.WU != 126855.22 {
		t.Errorf("Expected WU 126855.22, got %f", stats.WU)
	}
	if stats.Freq != 271.58 {
		t.Errorf("Expected Freq 271.58, got %f", stats.Freq)
	}
	if stats.MGHS != 9080.65 {
		t.Errorf("Expected MGHS 9080.65, got %f", stats.MGHS)
	}

	// Test additional metrics
	if stats.TA != 160 {
		t.Errorf("Expected TA 160, got %d", stats.TA)
	}
	if stats.Core != "A3197S" {
		t.Errorf("Expected Core 'A3197S', got '%s'", stats.Core)
	}
	if stats.BIN != 48 {
		t.Errorf("Expected BIN 48, got %d", stats.BIN)
	}
	if stats.PING != 36 {
		t.Errorf("Expected PING 36, got %d", stats.PING)
	}
	if stats.SoftOFF != 4 {
		t.Errorf("Expected SoftOFF 4, got %d", stats.SoftOFF)
	}
	if stats.ECHU != 0 {
		t.Errorf("Expected ECHU 0, got %d", stats.ECHU)
	}
	if stats.ECMM != 0 {
		t.Errorf("Expected ECMM 0, got %d", stats.ECMM)
	}

	// Test PLL0 array
	expectedPLL0 := []int{13026, 6474, 3643, 1177}
	if len(stats.PLL0) != len(expectedPLL0) {
		t.Errorf("Expected %d PLL0 entries, got %d", len(expectedPLL0), len(stats.PLL0))
	}
	for i, expected := range expectedPLL0 {
		if i < len(stats.PLL0) && stats.PLL0[i] != expected {
			t.Errorf("Expected PLL0[%d] %d, got %d", i, expected, stats.PLL0[i])
		}
	}

	// Test SF0 array
	expectedSF0 := []int{258, 276, 297, 318}
	if len(stats.SF0) != len(expectedSF0) {
		t.Errorf("Expected %d SF0 entries, got %d", len(expectedSF0), len(stats.SF0))
	}
	for i, expected := range expectedSF0 {
		if i < len(stats.SF0) && stats.SF0[i] != expected {
			t.Errorf("Expected SF0[%d] %d, got %d", i, expected, stats.SF0[i])
		}
	}

	// Test error counters
	if stats.CRC != 0 {
		t.Errorf("Expected CRC 0, got %d", stats.CRC)
	}
	if stats.COMCRC != 0 {
		t.Errorf("Expected COMCRC 0, got %d", stats.COMCRC)
	}

	// Test ATA0
	if stats.ATA0 != "800-65-2335-258-20" {
		t.Errorf("Expected ATA0 '800-65-2335-258-20', got '%s'", stats.ATA0)
	}

	// Test configuration fields
	if stats.LcdOnoff != 1 {
		t.Errorf("Expected LcdOnoff 1, got %d", stats.LcdOnoff)
	}
	if stats.Activation != 1 {
		t.Errorf("Expected Activation 1, got %d", stats.Activation)
	}
	if stats.WorkMode != 0 {
		t.Errorf("Expected WorkMode 0, got %d", stats.WorkMode)
	}
	if stats.WorkLevel != 0 {
		t.Errorf("Expected WorkLevel 0, got %d", stats.WorkLevel)
	}
	if stats.MPO != 800 {
		t.Errorf("Expected MPO 800, got %d", stats.MPO)
	}
	if stats.CALIALL != 7 {
		t.Errorf("Expected CALIALL 7, got %d", stats.CALIALL)
	}
	if stats.ADJ != 1 {
		t.Errorf("Expected ADJ 1, got %d", stats.ADJ)
	}
	if stats.NonceMask != 25 {
		t.Errorf("Expected NonceMask 25, got %d", stats.NonceMask)
	}

	// Test ID field
	if liteStat.ID != 1 {
		t.Errorf("Expected ID 1, got %d", liteStat.ID)
	}
}
