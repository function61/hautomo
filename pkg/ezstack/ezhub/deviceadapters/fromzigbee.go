package deviceadapters

import (
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
	"github.com/function61/hautomo/pkg/ezstack/zcl"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

// updates attributes from ReportAttributesCommand/ZoneStatusChangeNotificationCommand/etc.
func ZclIncomingMessageToAttributes(
	message *zcl.ZclIncomingMessage,
	actx *hubtypes.AttrsCtx,
	device *ezstack.Device,
) error {
	adapter := AdapterForModel(device.Model)

	switch cmd := message.Data.Command.(type) {
	case *cluster.DefaultResponseCommand:
		return nil // no-op, because this was already synchronously handled
	case *cluster.ReportAttributesCommand:
		// AttributeReports contains attributes a Zigbee sensor wants to report to us,
		// e.g. Temperature=21.3 °C, Humidity=40 %
		for _, report := range cmd.AttributeReports {
			// this will copy the reported attribute value to device current attributes, along
			// with the current timestamp
			if err := AttributeReportToAttributes(
				report,
				message.ClusterID,
				device,
				adapter,
				actx,
			); err != nil {
				return err
			}
		}

		return nil
	default:
		if err := adapter.HandleCommand(cmd, actx); err != nil {
			if err == errUnhandledCommand {
				return fmt.Errorf("unsupported command; got %s", spew.Sdump(message.Data.Command))
			} else {
				return err
			}
		}

		return nil
	}
}

// parses Zigbee AttributeReport message into structured sane attributes, taking into account
// manufacturer-specific "peculiarities"...
func AttributeReportToAttributes(
	report *cluster.AttributeReport,
	clusterId cluster.ClusterId,
	dev *ezstack.Device,
	adapter Adapter,
	actx *hubtypes.AttrsCtx,
) error {
	rxAttributeId := cluster.AttributeId(report.AttributeID)

	attrReport := report.Attribute // shorthand

	clDefinition := cluster.FindDefinition(clusterId)
	if clDefinition == nil {
		return fmt.Errorf("unknown cluster: %d", clusterId)
	}

	// TODO: clarify docs that definition datatypes are mainly needed for TX?
	attrDef := clDefinition.Attribute(rxAttributeId)

	// validate that received attr datatype is same as in definition. this is strictly not
	// necessary, as we can still meaningfully decode the value, but is a good sanity check
	// if this doesn't break any real-world use cases
	if attrDef != nil && attrDef.Type != attrReport.DataType {
		return fmt.Errorf(
			"device %s sent %s.%s with ZCL type %d (expected %d)",
			dev.NetworkAddress,
			clDefinition.Name(),
			attrDef.Name,
			attrReport.DataType,
			attrDef.Type)
	}

	// "genBasic.unknown(65281)" | "genBasic.modelId"
	keyDisplay := func() string {
		if attrDef == nil {
			return fmt.Sprintf("%s.unknown(%d)", clDefinition.Name(), report.AttributeID)
		} else {
			return fmt.Sprintf("%s.%s", clDefinition.Name(), attrDef.Name)
		}
	}()

	if parser := adapter.ParserForAttribute(keyDisplay); parser != nil {
		return parser(attrReport, actx)
	} else {
		val := cluster.SerializeMagicValue(attrReport.DataType, attrReport.Value)

		log.Printf(
			"AttributeReport from %s: unimplemented cluster %s (value %s)",
			dev.NetworkAddress,
			keyDisplay,
			val)

		return nil
	}
}
