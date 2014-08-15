package main

import (
	"fmt"
	"net"
	"os"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	Host           string             `long:"host" description:"host" default:"127.0.0.1"`
	Port           uint               `long:"port" description:"port" default:"10500"`
	Domain         string             `long:"domain" description:"domain" required:"true"`
	DBPath         string             `long:"db-path" description:"db path for sqlite3" default:":memory:"`
	AllowIPFunc    func(string) error `long:"allow-ip" description:"allow ip list (e.g. 127.0.0.1)"`
	AllowIPNetFunc func(string) error `long:"allow-ipnet" description:"allow ipnet list (e.g. 192.168.0.1/24)"`
	AllowHost      []string           `long:"allow-host" description:"allow host list (e.g. localhost)"`
	AllowIP        []string
	AllowIPNet     []*net.IPNet
}

var opts Options

func init() {
	opts.AllowIP = []string{}
	opts.AllowIPNet = []*net.IPNet{}
	opts.AllowIPFunc = func(ip string) error {
		if net.ParseIP(ip) == nil {
			return fmt.Errorf("invalid ip: %s", ip)
		}
		opts.AllowIP = append(opts.AllowIP, ip)
		return nil
	}
	opts.AllowIPNetFunc = func(ipnet string) error {
		_, ipNet, err := net.ParseCIDR(ipnet)
		if err != nil {
			return err
		}
		opts.AllowIPNet = append(opts.AllowIPNet, ipNet)
		return nil
	}

	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "tenco"
	parser.Usage = "[OPTIONS]"
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
	fmt.Printf("[tenco]\n")
	fmt.Printf("Host       : %s\n", opts.Host)
	fmt.Printf("Port       : %d\n", opts.Port)
	fmt.Printf("Domain     : %s\n", opts.Domain)
	fmt.Printf("DBPath     : %s\n", opts.DBPath)
	for i, v := range opts.AllowIP {
		if i == 0 {
			fmt.Printf("AllowIP    : %s\n", v)
		} else {
			fmt.Printf("             %s\n", v)
		}
	}
	for i, v := range opts.AllowIPNet {
		if i == 0 {
			fmt.Printf("AllowIPNet : %s\n", v.String())
		} else {
			fmt.Printf("             %s\n", v.String())
		}
	}
	for i, v := range opts.AllowHost {
		if i == 0 {
			fmt.Printf("AllowHost  : %s\n", v)
		} else {
			fmt.Printf("             %s\n", v)
		}
	}
}
