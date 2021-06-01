package coordinator

// commands one can ask the ZNP to do

import (
	"fmt"
	"time"

	"github.com/function61/hautomo/pkg/ezstack/zigbee"
	"github.com/function61/hautomo/pkg/ezstack/znp"
)

func (c *Coordinator) Reset() error {
	if _, err := c.syncRequestResponseRetryable(func() error {
		c.networkProcessor.SysResetReq(1) // we don't know about TX errors for async requests

		return nil
	}, SysResetIndType, 15*time.Second, 5); err != nil {
		return fmt.Errorf("SysResetReq: %w", err)
	}

	return nil
}

func (c *Coordinator) ActiveEndpoints(nwkAddress string) (*znp.ZdoActiveEpRsp, error) {
	req := func() error {
		status, err := c.networkProcessor.ZdoActiveEpReq(nwkAddress, nwkAddress)
		if err := firstError(err, func() error { return status.Status.Error() }); err != nil {
			return fmt.Errorf("unable to request active endpoints: %w", err)
		}
		return nil
	}

	response, err := c.syncRequestResponseRetryable(req, ZdoActiveEpRspType, defaultTimeout, 3)
	if err == nil {
		return response.(*znp.ZdoActiveEpRsp), nil
	}
	return nil, err
}

func (c *Coordinator) NodeDescription(nwkAddress string) (*znp.ZdoNodeDescRsp, error) {
	req := func() error {
		status, err := c.networkProcessor.ZdoNodeDescReq(nwkAddress, nwkAddress)
		if err := firstError(err, func() error { return status.Status.Error() }); err != nil {
			return fmt.Errorf("unable to request node description: %w", err)
		}
		return nil
	}

	response, err := c.syncRequestResponseRetryable(req, ZdoNodeDescRspType, defaultTimeout, 3)
	if err == nil {
		return response.(*znp.ZdoNodeDescRsp), nil
	}
	return nil, err
}

func (c *Coordinator) SimpleDescription(nwkAddress string, endpoint zigbee.EndpointId) (*znp.ZdoSimpleDescRsp, error) {
	req := func() error {
		status, err := c.networkProcessor.ZdoSimpleDescReq(nwkAddress, nwkAddress, endpoint)
		if err := firstError(err, func() error { return status.Status.Error() }); err != nil {
			return fmt.Errorf("unable to request simple description: %w", err)
		}
		return nil
	}

	response, err := c.syncRequestResponseRetryable(req, ZdoSimpleDescRspType, defaultTimeout, 3)
	if err == nil {
		return response.(*znp.ZdoSimpleDescRsp), nil
	}
	return nil, err
}

func (c *Coordinator) Bind(dstAddr string, srcAddress string, srcEndpoint zigbee.EndpointId, clusterId uint16,
	dstAddrMode znp.AddrMode, dstAddress string, dstEndpoint zigbee.EndpointId) (*znp.ZdoBindRsp, error) {
	req := func() error {
		status, err := c.networkProcessor.ZdoBindReq(dstAddr, srcAddress, srcEndpoint, clusterId, dstAddrMode, dstAddress, dstEndpoint)
		if err := firstError(err, func() error { return status.Status.Error() }); err != nil {
			return fmt.Errorf("unable to bind: %w", err)
		}
		return nil
	}

	response, err := c.syncRequestResponseRetryable(req, ZdoBindRspType, defaultTimeout, 3)
	if err == nil {
		return response.(*znp.ZdoBindRsp), nil
	}
	return nil, err
}

func (c *Coordinator) Unbind(dstAddr string, srcAddress string, srcEndpoint zigbee.EndpointId, clusterId uint16,
	dstAddrMode znp.AddrMode, dstAddress string, dstEndpoint zigbee.EndpointId) (*znp.ZdoUnbindRsp, error) {
	req := func() error {
		status, err := c.networkProcessor.ZdoUnbindReq(dstAddr, srcAddress, srcEndpoint, clusterId, dstAddrMode, dstAddress, dstEndpoint)
		if err := firstError(err, func() error { return status.Status.Error() }); err != nil {
			return fmt.Errorf("unable to unbind: %w", err)
		}
		return nil
	}

	response, err := c.syncRequestResponseRetryable(req, ZdoUnbindRspType, defaultTimeout, 3)
	if err == nil {
		return response.(*znp.ZdoUnbindRsp), nil
	}
	return nil, err
}

func (c *Coordinator) DataRequest(dstAddr string, dstEndpoint zigbee.EndpointId, srcEndpoint zigbee.EndpointId, clusterId uint16, options *znp.AfDataRequestOptions, radius uint8, data []uint8) (*znp.AfIncomingMessage, error) {
	req := func(networkAddress string, transactionId uint8) error {
		status, err := c.networkProcessor.AfDataRequest(networkAddress, dstEndpoint, srcEndpoint, clusterId, transactionId, options, radius, data)
		if err := firstError(err, func() error { return status.Status.Error() }); err != nil {
			return fmt.Errorf("DataRequest: %w", err)
		}
		return nil
	}

	return c.syncDataRequestResponseRetryable(req, dstAddr, nextTransactionId(), defaultTimeout, 3)
}

type errFn func() error

func firstError(err1 error, err2 func() error) error {
	if err1 != nil {
		return err1
	}

	return err2()
}
