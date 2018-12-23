package cnn

import (
	"github.com/dereklstinson/GoCuNets/cudnn"
	"github.com/dereklstinson/GoCuNets/layers"
	gocudnn "github.com/dereklstinson/GoCudnn"
)

//MakeOutputTensor makes the output tensor of the layer
func (c *Layer) MakeOutputTensor(handle *cudnn.Handler, input *layers.IO) (*layers.IO, error) {
	dims, err := c.conv.OutputDim(input.T(), c.w.T())
	if err != nil {
		return nil, err
	}
	frmt, dtype, _, err := c.w.Properties()
	if err != nil {
		return nil, err
	}
	managedmem := c.w.IsManaged()
	output, err := layers.BuildIO(frmt, dtype, dims, managedmem)
	if err != nil {
		return nil, err
	}
	return output, nil
}

//SetBestAlgosConsidering this method will set the best algos for the fwd, bwddata, and bwdfilter algos. and return the workspace size along with an error
//if an error is found the function will not set any values,
//Here are some simple rules to the function
//if fastest is marked true. Then it will find the fastest algo no mater what worksize is.
//if fastest is set to false. It will check if wspace is greater than zero then it will set the algos to the fastest algo considering the workspace size, and return the largest wspacesize in all the algos
//else it will find and set the fastest algos with no workspace size and return 0
func (c *Layer) SetBestAlgosConsidering(handle *cudnn.Handler, x, y *layers.IO, wspacelimit int, fastest bool) (gocudnn.SizeT, error) {
	return c.conv.SetBestAlgosConsidering(handle, x.T(), y.T(), c.w.T(), wspacelimit, fastest)
}

//SetBestAlgosConsideringDims4d this method will set the best algos for the fwd, bwddata, and bwdfilter algos. and return the workspace size along with an error
//if an error is found the function will not set any values,
//Here are some simple rules to the function
//if fastest is marked true. Then it will find the fastest algo no mater what worksize is.
//if fastest is set to false. It will check if wspace is greater than zero then it will set the algos to the fastest algo considering the workspace size, and return the largest wspacesize in all the algos
//else it will find and set the fastest algos with no workspace size and return 0
func (c *Layer) SetBestAlgosConsideringDims4d(handle *cudnn.Handler, x, y, w []int32, wspacelimit int, fastest bool) (gocudnn.SizeT, error) {
	frmt, data, _, err := c.w.Properties()
	if err != nil {
		return 0, err
	}
	return c.conv.SetBestAlgosConsideringDims4d(handle, x, y, w, wspacelimit, fastest, data, frmt)
}

//FilterProps returns the filter properties of the Convolution Layer
func (c *Layer) FilterProps() (cudnn.TensorFormat, cudnn.DataType, []int32, error) {
	return c.w.Properties()
}