package cnntranspose

import (
	"errors"

	"github.com/dereklstinson/GoCuNets/cudnn"
	"github.com/dereklstinson/GoCuNets/cudnn/reshapes"
	"github.com/dereklstinson/GoCuNets/layers"
	"github.com/dereklstinson/GoCuNets/layers/cnn"
	"github.com/dereklstinson/GoCuNets/utils"
	gocudnn "github.com/dereklstinson/GoCudnn"
)

//S2B does the shape2batch method of conv transform
func S2B(handle *cudnn.Handler,
	frmt cudnn.TensorFormat,
	dtype cudnn.DataType,
	window []int32, // window is the window of the shape2batch.
	filterdims []int32,
	convmode gocudnn.ConvolutionMode,
	pad,
	stride,
	dilation []int32,
	managedmem bool) (*Layer, error) {
	conv, err := cnn.SetupDynamic(handle, frmt, dtype, filterdims, filterdims, convmode, pad, stride, dilation, managedmem)
	if err != nil {
		return nil, err
	}
	reshaper, err := reshapes.Stage(handle)
	if err != nil {
		return nil, err
	}

	return &Layer{
		conv:      conv,
		mode:      convtransposes2b,
		trans:     reshaper,
		s2bwindow: window,
	}, nil
}

func (l *Layer) s2bforward(handle *cudnn.Handler, wspace *gocudnn.Malloced, x, y *layers.IO) error {
	frmt, dtype, dims, managed, err := l.trans.GetS2BOutputProperties(handle, x.T(), l.s2bwindow)
	if l.previouss2b == nil {
		l.hiddenmem, err = layers.BuildIO(frmt, dtype, dims, managed)
		if err != nil {
			return err
		}
		l.hiddenmem2, err = layers.BuildIO(frmt, dtype, l.conv.OutputDims(dims), managed)
		if err != nil {
			return err
		}
		l.previouss2b = dims
	} else if utils.CompareInt32(dims, l.previouss2b) == false {
		err = l.hiddenmem.Destroy()
		if err != nil {
			return err
		}
		err = l.hiddenmem2.Destroy()
		if err != nil {
			return err
		}
		l.hiddenmem, err = layers.BuildIO(frmt, dtype, dims, managed)
		if err != nil {
			return err
		}
		l.hiddenmem2, err = layers.BuildIO(frmt, dtype, l.conv.OutputDims(dims), managed)
		if err != nil {
			return err
		}
		l.previouss2b = dims
	}
	err = handle.Sync()
	if err != nil {
		return err
	}
	err = l.trans.S2BForward(handle, x.T(), l.hiddenmem.T())
	if err != nil {
		return err
	}
	err = handle.Sync()
	if err != nil {
		return err
	}
	err = l.conv.ForwardProp(handle, wspace, l.hiddenmem, l.hiddenmem2)
	if err != nil {
		return err
	}
	err = handle.Sync()
	if err != nil {
		return err
	}
	return l.trans.B2SForward(handle, l.hiddenmem2.T(), y.T())

}

func (l *Layer) s2bBackPropFilterData(handle *cudnn.Handler, wspace *gocudnn.Malloced, x, y *layers.IO) error {
	err := handle.Sync()
	if err != nil {
		return err
	}
	err = l.trans.B2SBackward(handle, l.hiddenmem2.DeltaT(), y.DeltaT())
	if err != nil {
		return err
	}
	err = handle.Sync()
	if err != nil {
		return err
	}
	err = l.conv.BackPropFilterData(handle, wspace, l.hiddenmem, l.hiddenmem2)
	if err != nil {
		return err
	}
	err = handle.Sync()
	if err != nil {
		return err
	}
	return l.trans.S2BBackward(handle, x.DeltaT(), l.hiddenmem.DeltaT())

}
func (l *Layer) s2bBackPropData(handle *cudnn.Handler, wspace *gocudnn.Malloced, x, y *layers.IO) error {
	err := handle.Sync()
	if err != nil {
		return err
	}
	err = l.trans.B2SBackward(handle, l.hiddenmem2.DeltaT(), y.DeltaT())
	if err != nil {
		return err
	}
	err = handle.Sync()
	if err != nil {
		return err
	}
	err = l.conv.BackPropData(handle, wspace, l.hiddenmem, l.hiddenmem2)
	if err != nil {
		return err
	}
	err = handle.Sync()
	if err != nil {
		return err
	}
	return l.trans.S2BBackward(handle, x.DeltaT(), l.hiddenmem.DeltaT())

}
func (l *Layer) s2outputIO(handle *cudnn.Handler, input *layers.IO) (*layers.IO, error) {
	frmt, dtype, dims, n1n2, managed, err := l.trans.GetS2BOutputPropertiesPLUS(handle, input.T(), l.s2bwindow)
	if l.previouss2b == nil {

		l.hiddenmem, err = layers.BuildIO(frmt, dtype, dims, managed)
		if err != nil {
			return nil, err
		}
		l.hiddenmem2, err = layers.BuildIO(frmt, dtype, l.conv.OutputDims(dims), managed)
		if err != nil {
			return nil, err
		}
		l.previouss2b = dims
		err = handle.Sync()
		if err != nil {
			return nil, err
		}
		frmt, dtype, dims, managed, err = l.trans.GetB2SOutputProperties(handle, l.hiddenmem2.T(), n1n2)
		if err == nil {
			return nil, err
		}
		return layers.BuildIO(frmt, dtype, dims, managed)
	}
	return nil, errors.New("Layer already has hidden build")

}