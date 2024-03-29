/*
* Define Open Edge Device Kit sender
* Do preparation work(connect to mosquitto, init oedk, store DataSorceId and DataPointId)
* Send data
*/

package mindconnect

import (
	"encoding/json"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"strconv"
	"strings"
	"time"
)

type OedkconnectSender struct {
	client			MQTT.Client
	config 			*configuration
	GetId			bool
}

var LoggingClient logger.LoggingClient

// newMindConnectSender returns new oedk sender instance.
func NewOEDKConnectSender() (*OedkconnectSender, error) {
	config, err := LoadConfigFromFile()
	if err != nil {
		return nil, fmt.Errorf("Read Open Edge Device Kit configuration failed: %v", err)
	}

	opts := MQTT.NewClientOptions()
	broker := fmt.Sprintf("%s", GetBaseURL(config))
	opts.AddBroker(broker)
	opts.SetClientID(config.Broker.Publisher)
	opts.SetUsername(config.Broker.User)
	opts.SetPassword(config.Broker.Password)
	opts.SetAutoReconnect(false)

	return &OedkconnectSender{
		client:		MQTT.NewClient(opts),
		config:		config,
		GetId:		false,
	}, nil
}

func GetBaseURL(config *configuration) string {
	protocol := strings.ToLower(config.Broker.Protocol)
	address := config.Broker.Address
	port := strconv.Itoa(config.Broker.Port)
	baseUrl := protocol + "://" + address + ":" + port
	return baseUrl
}

// do some preparations
func (sender *OedkconnectSender) Prepare(lc logger.LoggingClient) {
	LoggingClient = lc
	// connect to mosquitto server(local mqtt broker)
	LoggingClient.Info("Connecting to mosquitto")
	token := sender.client.Connect()
	if token.Wait() && token.Error() != nil {
		LoggingClient.Warn(fmt.Sprintf("Could not connect to mosquitto, drop event. Error: %s", token.Error().Error()))
		return
	}
	LoggingClient.Info("Connected into mosquitto")

	// initialize agent iff client is connected for the first time
	if !sender.config.Oedk.IsInitialized {
		token = sender.client.Publish(INIT_TOPIC, 0, false, sender.config.Oedk.InitJson)
		if token.Wait() && token.Error() != nil {
			LoggingClient.Error(token.Error().Error())
		}

		go func() {
			// init failed?
			token = sender.client.Subscribe(INITINFO_TOPIC, 0, sender.HandleInitInfo)
			if token.Wait() && token.Error() != nil {
				LoggingClient.Error(token.Error().Error())
			}
			select {}
		}()
	}

	// subscribe agentruntime/monitoring/diagnostic/onboarding
	//           agentruntime/monitoring/diagnostic/connection
	go func() {
		token = sender.client.Subscribe(ONBOARDING_TOPIC, 0, sender.HandleOnBoardTopic)
		if token.Wait() && token.Error() != nil {
			LoggingClient.Error(token.Error().Error())
		}

		select {}
	}()

	// subscribe cloud/monitoring/update/configuration/{protocol}
	go func() {
		ProTopic := strings.Replace(CONFIGPROINFO_TOPIC, "{protocol}", sender.config.Oedk.Protocol, 1)
		token = sender.client.Subscribe(ProTopic, 0, sender.HandleProTopic)
		if token.Wait() && token.Error() != nil {
			LoggingClient.Error(token.Error().Error())
		}

		select {}
	}()

	if sender.config.DataSource.Id == "" {
		sender.GetId = false
		return
	}
	for _, dp := range sender.config.DataSource.DataPoint {
		if dp.Id != "" {
			sender.GetId = true
			LoggingClient.Info("[Open Edge Device Kit] Read Ids for uploading date from file")
			return
		}
	}
}

func SendToSender(sender *OedkconnectSender) func(*appcontext.Context, ...interface{}) (bool, interface{}) {
	return sender.Send
}

func (sender *OedkconnectSender) Send(edgexcontext *appcontext.Context, params ...interface{}) (continuePipeline bool, result interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, nil
	}

	if !sender.client.IsConnected() {
		LoggingClient.Info("Connecting to mosquitto")
		token := sender.client.Connect()
		token.Wait()
		if token.Error() != nil {

			return false, fmt.Errorf("Could not connect to mosquitto, drop event. Error: %s", token.Error().Error())
		}
		LoggingClient.Info("Connected into mosquitto")
	}
	if !sender.GetId {
		return false, fmt.Errorf("[Open Edge Device Kit] DataSourceId and DataPointId invalid!")
	}

	var index = 0 // the number of DataPointId that linked with data
	value := params[0].(models.Event).Readings[0].Value
	jsonTemplate := "[\n" +
		"  {\n" +
		"    \"timestamp\": \"{timeStamp}\",\n" +
		"    \"values\": [\n" +
		"      {\n" +
		"        \"dataPointId\": \"{dataPointId}\",\n" +
		"        \"value\": \"{dataValue}\",\n" +
		"        \"qualityCode\": \"{qualityCode}\"\n" +
		"      }\n" +
		"    ]\n" +
		"  }\n" +
		"]";
	jsonTemplate = strings.Replace(jsonTemplate, "{timeStamp}", time.Now().Format(time.RFC3339), 1)
	jsonTemplate = strings.Replace(jsonTemplate, "{dataPointId}", sender.config.DataSource.DataPoint[index].Id, 1)
	jsonTemplate = strings.Replace(jsonTemplate, "{dataValue}", value, 1)
	jsonTemplate = strings.Replace(jsonTemplate, "{qualityCode}", "0", 1)

	DataTopic := strings.Replace(UPLOADDATA_TOPIC, "{protocol}", "OPCUA", 1)
	DataTopic = strings.Replace(DataTopic, "{dataSourceId}", sender.config.DataSource.Id, 1)
	token := sender.client.Publish(DataTopic, 0, false, jsonTemplate)

	if token.Wait() && token.Error() != nil {
		return false, token.Error().Error()
	}
	edgexcontext.LoggingClient.Info("Sent data to MQTT Broker")
	edgexcontext.LoggingClient.Trace("Data exported", "Transport", "MQTT", clients.CorrelationHeader, edgexcontext.CorrelationID)
	err := edgexcontext.MarkAsPushed()
	if err != nil {
		edgexcontext.LoggingClient.Error(err.Error())
	}
	return true, nil
}

// transform Mqtt message to map struct
func (sender *OedkconnectSender) HandleInitInfo(client MQTT.Client, message MQTT.Message) {
	var response map[string]interface{}
	json.Unmarshal(message.Payload(), &response)

	value := response["value"].(float64)
	status := response["status"].(string)

	if value == 1 {
		LoggingClient.Error(fmt.Sprintf("[Open Edge Device Kit] %s", status))
		sender.config.Oedk.IsInitialized = false
		return
	}
	sender.config.Oedk.IsInitialized = true
	LoggingClient.Debug("[Open Edge Device Kit] Init Successful")

	// update config file
	if err := UpdateConfigToFile(sender.config); err != nil {
		LoggingClient.Warn(fmt.Sprintf("%v", err))
	}
}
// transform Mqtt message to map struct
func (sender *OedkconnectSender) HandleOnBoardTopic(client MQTT.Client, message MQTT.Message) {
	var response map[string]interface{}
	json.Unmarshal(message.Payload(), &response)

	value := response["value"].(float64)
	state := response["state"].(string)

	if value == 2{  // in progress
		LoggingClient.Info(fmt.Sprintf("[Open Edge Device Kit] %s", state))
		return
	}
	if value == 3 || value == 4 { // failed or offboarded
		LoggingClient.Error(fmt.Sprintf("[Open Edge Device Kit] %s", state))
		return
	}
	if value == 1 { // success
		LoggingClient.Info(fmt.Sprintf("[Open Edge Device Kit] %s", state))
		return
	}

}

// transform Mqtt message to map struct
func (sender *OedkconnectSender) HandleProTopic(client MQTT.Client, message MQTT.Message) {
	LoggingClient.Info("[Open Edge Device Kit] Get latest configuration from Mindsphere")
	var response map[string]interface{}
	json.Unmarshal(message.Payload(), &response)

	sender.config.DataSource.Id = response["dataSourceId"].(string) // DataSourceId

	DataPoint, ok := response["dataPoints"];
	if  !ok {
		LoggingClient.Error("[Open Edge Device Kit] Please add at least one DataPoint in your DataSource")
		return
	}

	// DatePoints array
	for i, DataPointItem := range DataPoint.([]interface{}) {
		DataPoint := DataPointItem.(map[string]interface{})
		sender.config.DataSource.DataPoint[i].Id = DataPoint["dataPointId"].(string)
	}
	sender.GetId = true
	LoggingClient.Info("[Open Edge Device Kit] Get GetId for uploading date")

	// update config file
	if err := UpdateConfigToFile(sender.config); err != nil {
		LoggingClient.Warn(fmt.Sprintf("%v", err))
	}
}

func CoerceType(param interface{}) ([]byte, error) {
	var data []byte
	var err error

	switch param.(type) {
	case string:
		input := param.(string)
		data = []byte(input)

	case []byte:
		data = param.([]byte)

	case json.Marshaler:
		marshaler := param.(json.Marshaler)
		data, err = marshaler.MarshalJSON()
		if err != nil {
			return nil, nil
		}

	default:
		return nil, nil
	}
	return data, nil
}