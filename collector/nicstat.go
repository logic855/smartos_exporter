// nicstat collector
// this will :
//  - call nicstat
//  - gather network metrics
//  - feed the collector

package collector

import (
	"os/exec"
	"strconv"
	"strings"
	// Prometheus Go toolset
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// GZMLAGUsageCollector declares the data type within the prometheus metrics
// package.
type GZMLAGUsageCollector struct {
	gzMLAGUsageRead  *prometheus.GaugeVec
	gzMLAGUsageWrite *prometheus.GaugeVec
}

// NewGZMLAGUsageExporter returns a newly allocated exporter GZMLAGUsageCollector.
// It exposes the network bandwidth used by the MLAG interface
func NewGZMLAGUsageExporter() (*GZMLAGUsageCollector, error) {
	return &GZMLAGUsageCollector{
		gzMLAGUsageRead: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_network_mlag_receive_kilobytes",
			Help: "MLAG (rge0) receive statistic in KBytes.",
		}, []string{"device"}),
		gzMLAGUsageWrite: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_network_mlag_transmit_kilobytes",
			Help: "MLAG (rge0) transmit statistic in KBytes.",
		}, []string{"device"}),
	}, nil
}

// Describe describes all the metrics.
func (e *GZMLAGUsageCollector) Describe(ch chan<- *prometheus.Desc) {
	e.gzMLAGUsageRead.Describe(ch)
	e.gzMLAGUsageWrite.Describe(ch)
}

// Collect fetches the stats.
func (e *GZMLAGUsageCollector) Collect(ch chan<- prometheus.Metric) {
	e.nicstat()
	e.gzMLAGUsageRead.Collect(ch)
	e.gzMLAGUsageWrite.Collect(ch)
}

func (e *GZMLAGUsageCollector) nicstat() {
	// XXX needs enhancement :
	// use of nicstat will wait 2 seconds in order to collect statistics
	out, eerr := exec.Command("nicstat", "-i", "rge0", "1", "2").Output()
	if eerr != nil {
		log.Errorf("error on executing nicstat: %v", eerr)
	}
	perr := e.parseNicstatOutput(string(out))
	if perr != nil {
		log.Errorf("error on parsing nicstat: %v", perr)
	}
}

func (e *GZMLAGUsageCollector) parseNicstatOutput(out string) error {
	outlines := strings.Split(out, "\n")
	l := len(outlines)
	for _, line := range outlines[2 : l-1] {
		parsedLine := strings.Fields(line)
		readKb, err := strconv.ParseFloat(parsedLine[2], 64)
		if err != nil {
			return err
		}
		writeKb, err := strconv.ParseFloat(parsedLine[3], 64)
		if err != nil {
			return err
		}
		e.gzMLAGUsageRead.With(prometheus.Labels{"device": "rge0"}).Set(readKb)
		e.gzMLAGUsageWrite.With(prometheus.Labels{"device": "rge0"}).Set(writeKb)
	}
	return nil
}
