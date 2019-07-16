package filter

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"io/ioutil"
)

// filter rulle
type Rule struct {
	Enable         bool   `json:"enable"`
	Device         string `json:"device,omitempty"`
	Parameter      string `json:"parameter,omitempty"`
	Operation      string `json:"operation,omitempty"`
	Operand        string `json:"operand,omitempty"`
	Type           string `json:"type,omitempty"`
	FilterResult   bool   `json:"filterResult"`
}

var ruleChanges = make(chan Rule, 2)
// empty rule
var rule Rule

var confDir = "./res/rule.toml"

// LoadRuleFromFile use to load toml ruleuration
func LoadRuleFromFile(LoggingClient logger.LoggingClient){
	// Read toml file
	file, err := ioutil.ReadFile(confDir)
	if err != nil {
		LoggingClient.Error("could not load rule file (%s): %v", confDir, err.Error())
		return
	}

	// reformat toml to rule struct as defined previously
	err = toml.Unmarshal(file, &rule)
	if err != nil {
		LoggingClient.Error("unable to parse rule file (%s): %v", confDir, err.Error())
		return
	}

	go ListenForRule(LoggingClient)
	return
}

// UpdateRuleFromFile use to store toml ruleuration
func UpdateRuleToFile(rule *Rule) error {
	// reformat rule to bytes buffer
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(rule); err != nil {
		return fmt.Errorf("unable to parse to ruleuration file (%s): %v", err.Error())
	}

	if err := ioutil.WriteFile(confDir, buf.Bytes(), 0666); err != nil {
		return fmt.Errorf("could not store ruleuration file (%s): %v", confDir, err.Error())
	}
	return  nil
}

func RefreshRule(update Rule) error{
	ruleChanges <- update
	if err := UpdateRuleToFile(&update); err != nil {
		return err
	}
	return nil
}

func ReturnRule() Rule {
	return rule
}

func ListenForRule(LoggingClient logger.LoggingClient) {
	for {
		select {
		case rule = <-ruleChanges:
			if rule.Enable {
				LoggingClient.Info(fmt.Sprintf("Rule added: %v.", rule))
			} else {
				LoggingClient.Info("Rule deleted.")
			}
		}
	}
}