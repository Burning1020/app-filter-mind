package filter

import (
	"fmt"
	"github.com/caibirdme/yql"
	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"strconv"
)

//
//XXXXXXXXXXXXXXXXXXXXXXXXxxx
func FilterByValue(edgexcontext *appcontext.Context, params ...interface{}) (continuePipeline bool, result interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, nil
	}

	reading := params[0].(models.Event).Readings[0]

	if rule.Enable {
		switch rule.Type {
		case "int":
			newVal, _ := strconv.Atoi(reading.Value)
			rawYQL := fmt.Sprintf(`device = '%s' and parameter = '%s' and value %s %s`,
				rule.Device, rule.Parameter, rule.Operation, rule.Operand)

			result, _ := yql.Match(rawYQL, map[string]interface{}{
				"device":    reading.Device,
				"parameter": reading.Name,
				"value":     newVal,
			})

			if result {
				edgexcontext.LoggingClient.Debug(fmt.Sprintf("%s  %v", reading.Value, rule.FilterResult))
				return rule.FilterResult, nil
			} else {
				edgexcontext.LoggingClient.Debug(fmt.Sprintf("%s  %v", reading.Value, !rule.FilterResult))
				return !rule.FilterResult, nil
			}

		case "float":
			newVal, _ := strconv.ParseFloat(reading.Value, 64)
			rawYQL := fmt.Sprintf(`device = '%s' and parameter = '%s' and value %s %s`,
				rule.Device, rule.Parameter, rule.Operation, rule.Operand)

			result, _ := yql.Match(rawYQL, map[string]interface{}{
				"device":    reading.Device,
				"parameter": reading.Name,
				"value":     newVal,
			})

			if result {
				edgexcontext.LoggingClient.Debug(fmt.Sprintf("%s  %v", reading.Value, rule.FilterResult))
				return rule.FilterResult, nil
			} else {
				edgexcontext.LoggingClient.Debug(fmt.Sprintf("%s  %v", reading.Value, !rule.FilterResult))
				return !rule.FilterResult, nil
			}
		}

	}

	edgexcontext.LoggingClient.Debug(fmt.Sprintf("value=%s,  filterResout=true", reading.Value))
	// export the data by default
	return true, nil
}


//// cast value to valueType type
//func newValue(value string, valueType string) (interface{}, error) {
//	castError := "fail to parse %v reading, %v"
//	switch valueType {
//	case "String":
//		return value, nil
//	case "Uint8":
//		value, err := cast.ToUint8E(value)
//		if err != nil {
//			return nil, fmt.Errorf(castError, value, valueType)
//		}
//		return value, nil
//	case "Uint16":
//		value, err := cast.ToUint16E(value)
//		if err != nil {
//			return nil, fmt.Errorf(castError, value, valueType)
//		}
//		return value, nil
//	case "Uint32":
//		value, err := cast.ToUint32E(value)
//		if err != nil {
//			return nil, fmt.Errorf(castError, value, valueType)
//		}
//		return value, nil
//	case "Uint64":
//		value, err := cast.ToUint64E(value)
//		if err != nil {
//			return nil, fmt.Errorf(castError, value, valueType)
//		}
//		return value, nil
//	case "Int8":
//		value, err := cast.ToInt8E(value)
//		if err != nil {
//			return nil, fmt.Errorf(castError, value, valueType)
//		}
//		return value, nil
//	case "Int16":
//		value, err := cast.ToInt16E(value)
//		if err != nil {
//			return nil, fmt.Errorf(castError, value, valueType)
//		}
//		return value, nil
//	case "Int32":
//		value, err := cast.ToInt32E(value)
//		if err != nil {
//			return nil, fmt.Errorf(castError, value, valueType)
//		}
//		return value, nil
//	case "Int64":
//		value, err := cast.ToInt64E(value)
//		if err != nil {
//			return nil, fmt.Errorf(castError, value, valueType)
//		}
//		return value, nil
//	case "Float32":
//		value, err := cast.ToFloat32E(value)
//		if err != nil {
//			return nil, fmt.Errorf(castError, value, valueType)
//		}
//		return value, nil
//	case "Float64":
//		value, err := cast.ToFloat64E(value)
//		if err != nil {
//			return nil, fmt.Errorf(castError, value, valueType)
//		}
//		return value, nil
//	default:
//		return nil, fmt.Errorf("return result fail, none supported value type: %v", valueType)
//	}
//}

/*
func getValueType(name string, lg logger.LoggingClient) (string, error) {

	url := "" + clients.ApiValueDescriptorRoute
	params := types.EndpointParams{
		ServiceKey:  clients.CoreDataServiceKey,
		Path:        clients.ApiValueDescriptorRoute,
		UseRegistry: false,
		Url:         url,
		Interval:    clients.ClientMonitorDefault}
	vdc := coredata.NewValueDescriptorClient(params, startup.Endpoint{RegistryClient: &registryClient})

	vd, _ := vdc.ValueDescriptorForName(name, context.Background())
	return vd.Type, nil
}*/