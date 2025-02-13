package cat

import (
	"database/sql"

	go_metrics "github.com/kentik/go-metrics"
	old_logger "github.com/kentik/golog/logger"
	"github.com/kentik/ktranslate/pkg/eggs/logger"

	"github.com/kentik/ktranslate/pkg/api"
	"github.com/kentik/ktranslate/pkg/cat/auth"
	"github.com/kentik/ktranslate/pkg/filter"
	"github.com/kentik/ktranslate/pkg/formats"
	"github.com/kentik/ktranslate/pkg/inputs/flow"
	"github.com/kentik/ktranslate/pkg/inputs/vpc"
	"github.com/kentik/ktranslate/pkg/kt"
	"github.com/kentik/ktranslate/pkg/maps"
	"github.com/kentik/ktranslate/pkg/rollup"
	"github.com/kentik/ktranslate/pkg/sinks"
	"github.com/kentik/ktranslate/pkg/sinks/kentik"
	"github.com/kentik/ktranslate/pkg/util/gopatricia/patricia"

	model "github.com/kentik/ktranslate/pkg/util/kflow2"

	"github.com/bmatsuo/lmdb-go/lmdb"
)

const (
	HttpHealthCheckPath         = "/check"
	HttpAlertInboundPath        = "/chf"
	HttpCompanyID               = "sid"
	HttpAlertKey                = "key"
	MaxProxyListenerBufferAlloc = 10 * 1024 * 1024 // 10MB
	MSG_KEY_PREFIX              = 80
	HttpSenderID                = "sender_id"
	APP_PROTOCOL_COL            = "app_protocol"
	UDR_TYPE_INT                = "int"
	UDR_TYPE_BIGINT             = "bigint"
	UDR_TYPE_STRING             = "string"
	UDR_TYPE                    = "application_type"
)

// Config configuration parameters used by activate service
type Config struct {
	Listen            string
	SslCertFile       string
	SslKeyFile        string
	MappingFile       string
	Code2Region       string
	Code2City         string
	Threads           int
	ThreadsInput      int
	MaxThreads        int
	Format            formats.Format
	FormatRollup      formats.Format
	Compression       kt.Compression
	MaxFlowPerMessage int
	RollupAndAlpha    bool
	UDRFile           string
	GeoMapping        string
	Asn4              string
	Asn6              string
	AsnName           string
	DnsResolver       string
	SampleRate        uint32
	MaxBeforeSample   int
	Auth              *auth.AuthConfig
	SNMPFile          string
	SNMPDisco         bool
	TagMapType        maps.Mapper
	Kentik            *kt.KentikConfig
	VpcSource         vpc.CloudSource
	FlowSource        flow.FlowSource
	LogTee            chan string
	MetricsChan       chan []*kt.JCHF
}

type KTranslate struct {
	log            logger.ContextL
	config         *Config
	registry       go_metrics.Registry
	metrics        *KKCMetric
	alphaChans     []chan *Flow
	jchfChans      []chan *kt.JCHF
	inputChan      chan []*kt.JCHF
	mapr           *CustomMapper
	udrMapr        *UDRMapper
	pgdb           *sql.DB
	msgsc          chan *kt.Output
	envCode2Region *lmdb.Env
	envCode2City   *lmdb.Env
	ol             *old_logger.Logger
	sinks          map[sinks.Sink]sinks.SinkImpl
	format         formats.Formatter
	formatRollup   formats.Formatter
	kentik         *kentik.KentikSink // This one gets special handling
	rollups        []rollup.Roller
	doRollups      bool
	filters        []filter.Filter
	geo            *patricia.GeoTrees
	asn            *patricia.Uint32Trees
	resolver       *Resolver
	auth           *auth.Server
	apic           *api.KentikApi
	tooBig         chan int
	tagMap         maps.TagMapper
	vpc            vpc.VpcImpl
	nfs            *flow.KentikDriver
}

type CustomMapper struct {
	Customs map[uint32]string `json:"customs"`
}

type UDR struct {
	ColumnName      string
	ApplicationName string
	Type            string
}

type UDRMapper struct {
	UDRs     map[int32]map[string]*UDR
	Subtypes map[string]map[string]*UDR
}

type hc struct {
	Flows          float64
	DroppedFlows   float64
	FlowsOut       float64
	Errors         float64
	AlphaQ         int64
	JCHFQ          int64
	AlphaQDrop     float64
	InputQ         float64
	InputQLen      float64
	Sinks          map[sinks.Sink]map[string]float64
	SnmpDeviceData map[string]map[string]float64
}

type Flow struct {
	CompanyId int
	CHF       model.CHF
}

type KKCMetric struct {
	Flows          go_metrics.Meter
	FlowsOut       go_metrics.Meter
	DroppedFlows   go_metrics.Meter
	Errors         go_metrics.Meter
	AlphaQ         go_metrics.Gauge
	JCHFQ          go_metrics.Gauge
	AlphaQDrop     go_metrics.Meter
	InputQ         go_metrics.Meter
	InputQLen      go_metrics.Meter
	SnmpDeviceData *kt.SnmpMetricSet
}
