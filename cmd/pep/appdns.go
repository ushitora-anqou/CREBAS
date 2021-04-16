package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/miekg/dns"
	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/naoki9911/CREBAS/pkg/capability"
	"github.com/naoki9911/CREBAS/pkg/netlinkext"
	"github.com/naoki9911/CREBAS/pkg/ofswitch"
)

func getQueryResultFromServer(r *dns.Msg) (*dns.Msg, error) {
	dnsClient := new(dns.Client)
	dnsClient.Net = "udp"
	response, _, err := dnsClient.Exchange(r, dnsServer)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func isDomainAllowed(caps capability.CapabilitySlice, domain string) bool {
	for _, cap := range caps {
		if cap.IsDomainAllowed(domain) {
			return true
		}

	}

	return false
}

func dnsHandler(w dns.ResponseWriter, r *dns.Msg) {
	switch r.Opcode {
	case dns.OpcodeQuery:
		handleDnsQuery(w, r)
	}

}

func handleDnsQuery(w dns.ResponseWriter, r *dns.Msg) {
	clientIP := net.ParseIP(strings.Split(w.RemoteAddr().String(), ":")[0])

	log.Printf("info: Client is %v", clientIP.String())

	selectedApps := apps.Where(func(a app.AppInterface) bool {
		links := a.Links().Where(func(l *netlinkext.LinkExt) bool {
			return l.Addr.IP.String() == clientIP.String()
		})

		return len(links) != 0
	})

	if len(selectedApps) != 1 {
		log.Printf("info: Client %v not found", clientIP.String())
		return
	}

	selectedApp := selectedApps[0]
	caps := selectedApp.Capabilities().Where(func(c *capability.Capability) bool {
		return c.CapabilityName == capability.CAPABILITY_NAME_EXTERNAL_COMMUNICATION
	})

	var queryToServer []dns.Question
	var RRs []dns.RR
	for _, question := range r.Question {

		// Checks only A or AAAA Record
		if question.Qtype != dns.TypeA && question.Qtype != dns.TypeAAAA {
			queryToServer = append(queryToServer, question)
			continue
		}
		domain := question.Name[:len(question.Name)-1]
		if isDomainAllowed(caps, domain) {
			queryToServer = append(queryToServer, question)
		}
	}

	if len(queryToServer) != 0 {
		r.Question = queryToServer
		resFromServer, err := getQueryResultFromServer(r)
		if err != nil {
			log.Printf("error: Failed to lookup %v", err.Error())
		}
		RRs = append(RRs, resFromServer.Answer...)
	}

	response := &dns.Msg{}
	response.SetReply(r)
	response.Answer = RRs

	for _, res := range response.Answer {
		fmt.Println(res)
	}

	w.WriteMsg(response)
}

func startDNSServer(s *ofswitch.OFSwitch) {

	interfaceIP := s.Link.Addr.IP.String()
	host := interfaceIP + ":53"

	dns.HandleFunc(".", dnsHandler)

	server := &dns.Server{Addr: host, Net: "udp"}
	log.Printf("info: Starting at %s\n", host)
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("error: Failed to start server: %s\n ", err.Error())
	}
}
