package cnntranspose

import (
	"fmt"

	"github.com/dereklstinson/GoCuNets/devices/gpu/nvidia"
	"github.com/dereklstinson/GoCuNets/devices/gpu/nvidia/cudnn"
	"github.com/dereklstinson/GoCuNets/devices/gpu/nvidia/cudnn/deconvolution"
	"github.com/dereklstinson/GoCuNets/layers"
	gocudnn "github.com/dereklstinson/GoCudnn"
)

//MakeOutputTensor makes the output tensor of the layer
func (c *Layer) MakeOutputTensor(handle *cudnn.Handler, input *layers.Tensor) (*layers.Tensor, error) {
	dims, err := c.conv.OutputDim(input.Volume, c.w.Volume)
	if err != nil {
		fmt.Println(input.Properties())

		fmt.Println(c.w.Properties())
		return nil, err
	}
	frmt, dtype, _, err := c.w.Properties()
	if err != nil {
		return nil, err
	}

	output, err := layers.CreateTensor(handle, frmt, dtype, dims)
	if err != nil {
		return nil, err
	}
	return output, nil
}

//MakeOutputTensorInference makes the output tensor of the layer
func (c *Layer) MakeOutputTensorInference(handle *cudnn.Handler, input *layers.Tensor) (*layers.Tensor, error) {
	dims, err := c.conv.OutputDim(input.Volume, c.w.Volume)
	if err != nil {
		return nil, err
	}
	frmt, dtype, _, err := c.w.Properties()
	if err != nil {
		return nil, err
	}

	output, err := layers.CreateTensor(handle, frmt, dtype, dims)
	if err != nil {
		return nil, err
	}
	return output, nil
}

//FindOutputDims finds the outputdims fo the cnn
func (c *Layer) FindOutputDims(input *layers.Tensor) (dims []int32, err error) {
	dims, err = c.conv.OutputDim(input.Volume, c.w.Volume)
	return dims, err
}

//SetBestAlgosConsidering this method will set the best algos for the fwd, bwddata, and bwdfilter algos. and return the workspace size along with an error
//if an error is found the function will not set any values,
//Here are some simple rules to the function
//if fastest is marked true. Then it will find the fastest algo no mater what worksize is.
//if fastest is set to false. It will check if wspace is greater than zero then it will set the algos to the fastest algo considering the workspace size, and return the largest wspacesize in all the algos
//else it will find and set the fastest algos with no workspace size and return 0
func (c *Layer) SetBestAlgosConsidering(handle *cudnn.Handler, x, y *layers.Tensor, wspacelimit int, fastest bool) (uint, error) {
	return c.conv.SetBestAlgosConsidering(handle, x.Volume, y.Volume, c.w.Volume, wspacelimit, fastest)
}

//SetBestAlgosConsideringDims4d this method will set the best algos for the fwd, bwddata, and bwdfilter algos. and return the workspace size along with an error
//if an error is found the function will not set any values,
//Here are some simple rules to the function
//if fastest is marked true. Then it will find the fastest algo no mater what worksize is.
//if fastest is set to false. It will check if wspace is greater than zero then it will set the algos to the fastest algo considering the workspace size, and return the largest wspacesize in all the algos
//else it will find and set the fastest algos with no workspace size and return 0
func (c *Layer) SetBestAlgosConsideringDims4d(handle *cudnn.Handler, x, y, w []int32, wspacelimit int, fastest bool) (uint, error) {
	frmt, data, _, err := c.w.Properties()
	if err != nil {
		return 0, err
	}
	return c.conv.SetBestAlgosConsideringDims4d(handle, x, y, w, wspacelimit, fastest, data, frmt)
}

//FilterProps returns the filter properties of the DeConvolution Layer
func (c *Layer) FilterProps() (gocudnn.TensorFormat, gocudnn.DataType, []int32, error) {
	return c.w.Properties()
}

//GetFwdAlgoPerfList gets a list of forward performance algos
func (c *Layer) GetFwdAlgoPerfList(handle *cudnn.Handler, x, y *layers.Tensor, workspace *nvidia.Malloced) ([]deconvolution.ForwardPerformance, error) {
	return c.conv.GetFwdAlgoPerfList(handle, x.Volume, c.w.Volume, y.Volume, workspace)
}

//GetBwdDataAlgoPerfList gets a list of backward performance algos
func (c *Layer) GetBwdDataAlgoPerfList(handle *cudnn.Handler, x, y *layers.Tensor, workspace *nvidia.Malloced) ([]deconvolution.BackDataPerformance, error) {
	return c.conv.GetBwdDataAlgoPerfList(handle, x.Volume, c.w.Volume, y.Volume, workspace)
}

//GetBwdFiltAlgoPerfList gets a list of forward performance algos
func (c *Layer) GetBwdFiltAlgoPerfList(handle *cudnn.Handler, x, y *layers.Tensor, workspace *nvidia.Malloced) ([]deconvolution.BackFilterPerformance, error) {
	return c.conv.GetBwdFiltAlgoPerfList(handle, x.Volume, c.w.Volume, y.Volume, workspace)
}

//SetFwdAlgoPerformance sets the Performance Values
func (c *Layer) SetFwdAlgoPerformance(fwd deconvolution.ForwardPerformance) {
	c.conv.SetFwdPerformanceAlgo(fwd)
}

//SetBwdFiltAlgoPerformance sets the Performance Values
func (c *Layer) SetBwdFiltAlgoPerformance(bwdfilt deconvolution.BackFilterPerformance) {
	c.conv.SetBwdFiltPerformanceAlgo(bwdfilt)
}

//SetBwdDataAlgoPerformance sets the Performance Values
func (c *Layer) SetBwdDataAlgoPerformance(bwddata deconvolution.BackDataPerformance) {
	c.conv.SetBwdDataPerformanceAlgo(bwddata)
}
