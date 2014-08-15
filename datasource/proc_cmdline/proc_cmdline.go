package proc_cmdline

import (
	"errors"
	"io/ioutil"
	"log"
	"strings"

	"github.com/coreos/coreos-cloudinit/pkg"
)

const (
	ProcCmdlineLocation        = "/proc/cmdline"
	ProcCmdlineCloudConfigFlag = "cloud-config-url"
)

type procCmdline struct {
	Location string
}

func NewDatasource() *procCmdline {
	return &procCmdline{Location: ProcCmdlineLocation}
}

func (c *procCmdline) IsAvailable() bool {
	contents, err := ioutil.ReadFile(c.Location)
	if err != nil {
		return false
	}

	cmdline := strings.TrimSpace(string(contents))
	_, err = findCloudConfigURL(cmdline)
	return (err == nil)
}

func (c *procCmdline) AvailabilityChanges() bool {
	return false
}

func (c *procCmdline) ConfigRoot() string {
	return ""
}

func (c *procCmdline) FetchMetadata() ([]byte, error) {
	return []byte{}, nil
}

func (c *procCmdline) FetchUserdata() ([]byte, error) {
	contents, err := ioutil.ReadFile(c.Location)
	if err != nil {
		return nil, err
	}

	cmdline := strings.TrimSpace(string(contents))
	url, err := findCloudConfigURL(cmdline)
	if err != nil {
		return nil, err
	}

	client := pkg.NewHttpClient()
	cfg, err := client.GetRetry(url)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *procCmdline) Type() string {
	return "proc-cmdline"
}

func findCloudConfigURL(input string) (url string, err error) {
	err = errors.New("cloud-config-url not found")
	for _, token := range strings.Split(input, " ") {
		parts := strings.SplitN(token, "=", 2)

		key := parts[0]
		key = strings.Replace(key, "_", "-", -1)

		if key != "cloud-config-url" {
			continue
		}

		if len(parts) != 2 {
			log.Printf("Found cloud-config-url in /proc/cmdline with no value, ignoring.")
			continue
		}

		url = parts[1]
		err = nil
	}

	return
}