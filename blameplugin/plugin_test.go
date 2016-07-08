package blameplugin_test

import (
	"bytes"
	"io/ioutil"
	"strings"

	"github.com/cloudfoundry/cli/plugin/pluginfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/diego-blame/blameplugin"
	"fmt"
)

var _ = Describe("DiegoBlame", func() {
	Describe("Given a DiegoBlame object", func() {
		Context("when calling Run w/ the wrong args", func() {
			var plgn *DiegoBlame
			BeforeEach(func() {
				plgn = new(DiegoBlame)
			})
			It("then it should panic and exit", func() {
				Ω(func() {
					cli := new(pluginfakes.FakeCliConnection)
					plgn.Run(cli, []string{})
				}).Should(Panic())
			})
		})

		Context("when calling Run w/ the proper args", func() {
			var plgn *DiegoBlame
			var b *bytes.Buffer
			BeforeEach(func() {
				b = new(bytes.Buffer)
				plgn = &DiegoBlame{
					Writer: b,
				}
			})
			It("then it should print a table with the correct columns", func() {
				cli := new(pluginfakes.FakeCliConnection)
				plgn.Run(cli, []string{"hi", "there"})
				for _, col := range []string{"app name/instance", "State", "Host:Port", "Org", "Space", "Disk-Usage", "Disk-Quota", "Mem-Usage", "Mem-Quota", "CPU-Usage", "Uptime", "URIs"} {
					Ω(strings.ToLower(b.String())).Should(ContainSubstring(strings.ToLower(col)))
				}
			})
			It("then it should print something", func() {
				cli := new(pluginfakes.FakeCliConnection)
				plgn.Run(cli, []string{"hi", "there"})
				Ω(b.String()).ShouldNot(BeEmpty())
			})
		})
	})

	Describe("Given CallAppsApi", func() {
		Context("when called with a valid endpoint and cli", func() {
			var guidArray []string
			BeforeEach(func() {
				cli := new(pluginfakes.FakeCliConnection)
				b, _ := ioutil.ReadFile("fixtures/app.json")
				cli.CliCommandWithoutTerminalOutputReturns([]string{string(b)}, nil)
				guidArray = CallAppsAPI("/v2/apps", cli)
			})

			It("should return a valid list of app guids", func() {
				Ω(len(guidArray)).Should(Equal(2))
				Ω(guidArray[0]).Should(Equal("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"))
				Ω(guidArray[1]).Should(Equal("yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy"))
			})
		})
	})

	Describe("Given CallStatsApi", func() {
		Context("when called with a valid guid and cli and selector and a host the app is running on", func() {
			var stats []AppStat
			BeforeEach(func() {
				cli := new(pluginfakes.FakeCliConnection)
				b, _ := ioutil.ReadFile("fixtures/stats.json")
				cli.CliCommandWithoutTerminalOutputReturns([]string{string(b)}, nil)
				stats = CallStatsAPI("xxx-xxx-xxxxx-xxxxx", cli, "192.xxx.x.255")
			})
			It("then it should return the stats for apps instances", func() {
				Ω(len(stats)).Should(Equal(2))
				namePrefixMatch := "cool-server"
				Ω(stats[0].Stats.Name).Should(HavePrefix(namePrefixMatch))
				Ω(stats[1].Stats.Name).Should(HavePrefix(namePrefixMatch))
			})
		})

		Context("when called with a valid guid and cli and selector and a host the app is NOT running on", func() {
			var stats []AppStat
			BeforeEach(func() {
				cli := new(pluginfakes.FakeCliConnection)
				b, _ := ioutil.ReadFile("fixtures/stats.json")
				cli.CliCommandWithoutTerminalOutputReturns([]string{string(b)}, nil)
				stats = CallStatsAPI("xxx-xxx-xxxxx-xxxxx", cli, "192.not.x.111")
			})
			It("then it should return no stats", func() {
				Ω(len(stats)).Should(Equal(0))
			})
		})
	})
	Describe("Given an array of AppStat", func() {
		var stats []AppStat
		var b *bytes.Buffer
		var cli = new(pluginfakes.FakeCliConnection)
		Context("when table with application statistics is rendered", func() {
			BeforeEach(func() {
				b = new(bytes.Buffer)
				cli := new(pluginfakes.FakeCliConnection)
				statFile, _ := ioutil.ReadFile("fixtures/stats.json")
				cli.CliCommandWithoutTerminalOutputReturns([]string{string(statFile)}, nil)
				stats = CallStatsAPI("xxx-xxx-xxxxx-xxxxx", cli, "192.xxx.x.255")

			})
			It("should print the table with stats data sorted by mem ratio desc", func() {
				PrettyColumnPrint(stats, cli, b)
				firstRatio := "0.08930087089538574"
				secondRatio := "0.078125"
				out := b.String()
				Ω(strings.Index(out, firstRatio)).Should(BeNumerically("<", strings.Index(out, secondRatio)))
			})
			It("should print empty table when no data is returned by stats call", func(){
				e := []AppStat{}
				PrettyColumnPrint(e, cli, b)
				out := b.String()
				foundData := strings.Index(out, "cool-server")
				Ω(foundData).Should(BeNumerically("==", -1))
			})
		})
	})
})
