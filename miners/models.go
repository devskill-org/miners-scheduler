package miners

// AvalonState represents the state of an Avalon miner
type AvalonState int

// AvalonWorkMode represents the work mode of an Avalon miner
type AvalonWorkMode int

const (
	AvalonStateRunning AvalonState = 0 // Running
	AvalonStateMining  AvalonState = 1 // Mining
	AvalonStateStandBy AvalonState = 2 // StandBy
)

const (
	AvalonEcoMode      AvalonWorkMode = 0 // Eco
	AvalonStandardMode AvalonWorkMode = 1 // Standard
	AvalonSuperMode    AvalonWorkMode = 2 // Super
)

// String returns the string representation of the AvalonState
func (s AvalonState) String() string {
	switch s {
	case AvalonStateRunning:
		return "Running"
	case AvalonStateMining:
		return "Mining"
	case AvalonStateStandBy:
		return "StandBy"
	default:
		return "Unknown"
	}
}

// IsValid returns true if the AvalonState is a valid state
func (s AvalonState) IsValid() bool {
	switch s {
	case AvalonStateRunning, AvalonStateMining, AvalonStateStandBy:
		return true
	default:
		return false
	}
}

type AvalonQCommand struct {
	Command string `json:"command"`
}

type AvalonQHost struct {
	Address          string
	Port             int
	Version          *AvalonQVersion
	LiteStatsHistory []*AvalonLiteStats
}

// AddLiteStats appends a new AvalonLiteStats to the history and keeps only the last 5 entries.
func (h *AvalonQHost) AddLiteStats(stats *AvalonLiteStats) {
	h.LiteStatsHistory = append(h.LiteStatsHistory, stats)
	if len(h.LiteStatsHistory) > 5 {
		h.LiteStatsHistory = h.LiteStatsHistory[len(h.LiteStatsHistory)-5:]
	}
}

// ResetLiteStats keeps only the latest stats in history.
func (h *AvalonQHost) ResetLiteStats() {
	if len(h.LiteStatsHistory) > 0 {
		h.LiteStatsHistory = h.LiteStatsHistory[len(h.LiteStatsHistory)-1:]
	}
}

type AvalonQVersion struct {
	Status  []StatusItem  `json:"STATUS"`
	Version []VersionItem `json:"VERSION"`
	ID      int           `json:"id"`
}

type StatusItem struct {
	Status      string `json:"STATUS"`
	When        int64  `json:"When"`
	Code        int    `json:"Code"`
	Msg         string `json:"Msg"`
	Description string `json:"Description"`
}

type VersionItem struct {
	CGMiner       string `json:"CGMiner"`
	API           string `json:"API"`
	Prod          string `json:"PROD"`
	Model         string `json:"MODEL"`
	HWType        string `json:"HWTYPE"`
	SWType        string `json:"SWTYPE"`
	LVersion      string `json:"LVERSION"`
	BVersion      string `json:"BVERSION"`
	CGVersion     string `json:"CGVERSION"`
	HBMcuVersion  string `json:"HBMCUVERSION"`
	FANMcuVersion string `json:"FANMCUVERSION"`
	DNA           string `json:"DNA"`
	MAC           string `json:"MAC"`
}

type AvalonQLiteStats struct {
	Status []StatusItem `json:"STATUS"`
	Stats  []StatsItem  `json:"STATS"`
	ID     int          `json:"id"`
}

type StatsItem struct {
	MMIDSummary *AvalonLiteStats `json:"-"`
}

type AvalonLiteStats struct {
	Ver          string      `json:"ver"`
	LVer         string      `json:"lver"`
	BVer         string      `json:"bver"`
	HashMcu0Ver  string      `json:"hash_mcu0_ver"`
	FanMcuVer    string      `json:"fan_mcu_ver"`
	CPU          string      `json:"cpu"`
	FW           string      `json:"fw"`
	DNA          string      `json:"dna"`
	State        AvalonState `json:"state"`
	MemFree      int         `json:"mem_free"`
	NetFail      []int64     `json:"net_fail"`
	SSID         string      `json:"ssid"`
	RSSI         int         `json:"rssi"`
	NetDevType   int         `json:"net_dev_type"`
	SystemStatus string      `json:"system_status"`
	Elapsed      int64       `json:"elapsed"`
	BootBy       string      `json:"boot_by"`
	LW           int64       `json:"lw"`
	MH           int         `json:"mh"`
	DHW          int         `json:"dhw"`
	HW           int         `json:"hw"`
	DH           string      `json:"dh"`
	ITemp        int         `json:"itemp"`
	HBITemp      int         `json:"hbi_temp"`
	HBOTemp      int         `json:"hbo_temp"`
	TMax         int         `json:"tmax"`
	TAvg         int         `json:"tavg"`
	TarT         int         `json:"tart"`
	Fan1         int         `json:"fan1"`
	Fan2         int         `json:"fan2"`
	Fan3         int         `json:"fan3"`
	Fan4         int         `json:"fan4"`
	FanR         int         `json:"fanr"`
	SoftOffTime  int64       `json:"soft_off_time"`
	SoftOnTime   int64       `json:"soft_on_time"`
	Filter       int         `json:"filter"`
	FanErr       int         `json:"fan_err"`
	SoloAllowed  int         `json:"solo_allowed"`
	PS           []int       `json:"ps"`
	PCOMM_E      int         `json:"pcomm_e"`
	GHSspd       float64     `json:"ghs_spd"`
	DHspd        string      `json:"dh_spd"`
	GHSmm        float64     `json:"ghs_mm"`
	GHSavg       float64     `json:"ghs_avg"`
	WU           float64     `json:"wu"`
	Freq         float64     `json:"freq"`
	MGHS         float64     `json:"mghs"`
	TA           int         `json:"ta"`
	Core         string      `json:"core"`
	BIN          int         `json:"bin"`
	PING         int         `json:"ping"`
	SoftOFF      int         `json:"soft_off"`
	ECHU         int         `json:"echu"`
	ECMM         int         `json:"ecmm"`
	PLL0         []int       `json:"pll0"`
	SF0          []int       `json:"sf0"`
	CRC          int         `json:"crc"`
	COMCRC       int         `json:"comcrc"`
	ATA0         string      `json:"ata0"`
	LcdOnoff     int         `json:"lcd_onoff"`
	Activation   int         `json:"activation"`
	WorkMode     int         `json:"work_mode"`
	WorkLevel    int         `json:"work_level"`
	MPO          int         `json:"mpo"`
	CALIALL      int         `json:"caliall"`
	ADJ          int         `json:"adj"`
	NonceMask    int         `json:"nonce_mask"`
}
