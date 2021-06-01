package cluster

import "fmt"

type ZclDataType uint8

const (
	ZclDataTypeNoData        ZclDataType = 0x00
	ZclDataTypeData8         ZclDataType = 0x08
	ZclDataTypeData16        ZclDataType = 0x09
	ZclDataTypeData24        ZclDataType = 0x0a
	ZclDataTypeData32        ZclDataType = 0x0b
	ZclDataTypeData40        ZclDataType = 0x0c
	ZclDataTypeData48        ZclDataType = 0x0d
	ZclDataTypeData56        ZclDataType = 0x0e
	ZclDataTypeData64        ZclDataType = 0x0f
	ZclDataTypeBoolean       ZclDataType = 0x10
	ZclDataTypeBitmap8       ZclDataType = 0x18
	ZclDataTypeBitmap16      ZclDataType = 0x19
	ZclDataTypeBitmap24      ZclDataType = 0x1a
	ZclDataTypeBitmap32      ZclDataType = 0x1b
	ZclDataTypeBitmap40      ZclDataType = 0x1c
	ZclDataTypeBitmap48      ZclDataType = 0x1d
	ZclDataTypeBitmap56      ZclDataType = 0x1e
	ZclDataTypeBitmap64      ZclDataType = 0x1f
	ZclDataTypeUint8         ZclDataType = 0x20
	ZclDataTypeUint16        ZclDataType = 0x21
	ZclDataTypeUint24        ZclDataType = 0x22
	ZclDataTypeUint32        ZclDataType = 0x23
	ZclDataTypeUint40        ZclDataType = 0x24
	ZclDataTypeUint48        ZclDataType = 0x25
	ZclDataTypeUint56        ZclDataType = 0x26
	ZclDataTypeUint64        ZclDataType = 0x27
	ZclDataTypeInt8          ZclDataType = 0x28
	ZclDataTypeInt16         ZclDataType = 0x29
	ZclDataTypeInt24         ZclDataType = 0x2a
	ZclDataTypeInt32         ZclDataType = 0x2b
	ZclDataTypeInt40         ZclDataType = 0x2c
	ZclDataTypeInt48         ZclDataType = 0x2d
	ZclDataTypeInt56         ZclDataType = 0x2e
	ZclDataTypeInt64         ZclDataType = 0x2f
	ZclDataTypeEnum8         ZclDataType = 0x30
	ZclDataTypeEnum16        ZclDataType = 0x31
	ZclDataTypeSemiPrec      ZclDataType = 0x38
	ZclDataTypeSinglePrec    ZclDataType = 0x39
	ZclDataTypeDoublePrec    ZclDataType = 0x3a
	ZclDataTypeOctetStr      ZclDataType = 0x41
	ZclDataTypeCharStr       ZclDataType = 0x42
	ZclDataTypeLongOctetStr  ZclDataType = 0x43
	ZclDataTypeLongCharStr   ZclDataType = 0x44
	ZclDataTypeArray         ZclDataType = 0x48
	ZclDataTypeStruct        ZclDataType = 0x4c
	ZclDataTypeSet           ZclDataType = 0x50
	ZclDataTypeBag           ZclDataType = 0x51
	ZclDataTypeTod           ZclDataType = 0xe0
	ZclDataTypeDate          ZclDataType = 0xe1
	ZclDataTypeUtc           ZclDataType = 0xe2
	ZclDataTypeClusterId     ZclDataType = 0xe8
	ZclDataTypeAttrId        ZclDataType = 0xe9
	ZclDataTypeBacOid        ZclDataType = 0xea
	ZclDataTypeIeeeAddr      ZclDataType = 0xf0
	ZclDataType_128BitSecKey ZclDataType = 0xf1
	ZclDataTypeUnknown       ZclDataType = 0xff
)

type ZclStatus uint8

const (
	ZclStatusSuccess ZclStatus = 0x00
	ZclStatusFailure ZclStatus = 0x01
	// 0x02-0x7D are reserved.
	ZclStatusNotAuthorized            ZclStatus = 0x7E
	ZclStatusMalformedCommand         ZclStatus = 0x80
	ZclStatusUnsupClusterCommand      ZclStatus = 0x81
	ZclStatusUnsupGeneralCommand      ZclStatus = 0x82
	ZclStatusUnsupManuClusterCommand  ZclStatus = 0x83
	ZclStatusUnsupManuGeneralCommand  ZclStatus = 0x84
	ZclStatusInvalidField             ZclStatus = 0x85
	ZclStatusUnsupportedAttribute     ZclStatus = 0x86
	ZclStatusInvalidValue             ZclStatus = 0x87
	ZclStatusReadOnly                 ZclStatus = 0x88
	ZclStatusInsufficientSpace        ZclStatus = 0x89
	ZclStatusDuplicateExists          ZclStatus = 0x8a
	ZclStatusNotFound                 ZclStatus = 0x8b
	ZclStatusUnreportableAttribute    ZclStatus = 0x8c
	ZclStatusInvalidDataType          ZclStatus = 0x8d
	ZclStatusInvalidSelector          ZclStatus = 0x8e
	ZclStatusWriteOnly                ZclStatus = 0x8f
	ZclStatusInconsistentStartupState ZclStatus = 0x90
	ZclStatusDefinedOutOfBand         ZclStatus = 0x91
	ZclStatusInconsistent             ZclStatus = 0x92
	ZclStatusActionDenied             ZclStatus = 0x93
	ZclStatusTimeout                  ZclStatus = 0x94
	ZclStatusAbort                    ZclStatus = 0x95
	ZclStatusInvalidImage             ZclStatus = 0x96
	ZclStatusWaitForData              ZclStatus = 0x97
	ZclStatusNoImageAvailable         ZclStatus = 0x98
	ZclStatusRequireMoreImage         ZclStatus = 0x99

	// 0xbd-bf are reserved.
	ZclStatusHardwareFailure  ZclStatus = 0xc0
	ZclStatusSoftwareFailure  ZclStatus = 0xc1
	ZclStatusCalibrationError ZclStatus = 0xc2
	// 0xc3-0xff are reserved.
	ZclStatusCmdHasRsp ZclStatus = 0xFF // Non-standard status (used for Default Rsp)
)

// converts ZCL status to an error
func (z ZclStatus) Error() error {
	if z == ZclStatusSuccess {
		return nil
	} else {
		return fmt.Errorf("ZCL error: %d", z)
	}
}

type ZclCommand uint8

const (
	ZclCommandReadAttributes                     ZclCommand = 0x00
	ZclCommandReadAttributesResponse             ZclCommand = 0x01
	ZclCommandWriteAttributes                    ZclCommand = 0x02
	ZclCommandWriteAttributesUndivided           ZclCommand = 0x03
	ZclCommandWriteAttributesResponse            ZclCommand = 0x04
	ZclCommandWriteAttributesNoResponse          ZclCommand = 0x05
	ZclCommandConfigureReporting                 ZclCommand = 0x06
	ZclCommandConfigureReportingResponse         ZclCommand = 0x07
	ZclCommandReadReportingConfiguration         ZclCommand = 0x08
	ZclCommandReadReportingConfigurationResponse ZclCommand = 0x09
	ZclCommandReportAttributes                   ZclCommand = 0x0a
	ZclCommandDefaultResponse                    ZclCommand = 0x0b
	ZclCommandDiscoverAttributes                 ZclCommand = 0x0c
	ZclCommandDiscoverAttributesResponse         ZclCommand = 0x0d
	ZclCommandReadAttributesStructured           ZclCommand = 0x0e
	ZclCommandWriteAttributesStructured          ZclCommand = 0x0f
	ZclCommandWriteAttributesStructuredResponse  ZclCommand = 0x10
	ZclCommandDiscoverCommandsReceived           ZclCommand = 0x11
	ZclCommandDiscoverCommandsReceivedResponse   ZclCommand = 0x12
	ZclCommandDiscoverCommandsGenerated          ZclCommand = 0x13
	ZclCommandDiscoverCommandsGeneratedResponse  ZclCommand = 0x14
	ZclCommandDiscoverAttributesExtended         ZclCommand = 0x15
	ZclCommandDiscoverAttributesExtendedResponse ZclCommand = 0x16
)

type ReportDirection uint8

const (
	ReportDirectionAttributeReported ReportDirection = 0x00
	ReportDirectionAttributeReceived ReportDirection = 0x01
)
