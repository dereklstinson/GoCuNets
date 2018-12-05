package batchnorm

import (
	"github.com/dereklstinson/GoCuNets/cudnn"
	"github.com/dereklstinson/GoCuNets/cudnn/batchnorm"
	"github.com/dereklstinson/GoCuNets/layers"
	gocudnn "github.com/dereklstinson/GoCudnn"
)

const alphaforwarddefault = 1
const betaforwarddefault = 1
const alphabackwarddefault = 1
const betabackwarddefault = 1
const alphabackwardparamdefault = 1
const betabackwardparamdefault = 1

//Layer the ops of a batch norm
type Layer struct {
	b       *batchnorm.Ops
	fw      abscalars
	bwp     abscalars
	bwd     abscalars
	eps     float64
	af      float64
	counter uint64
	mode    gocudnn.BatchNormMode
	managed bool
}
type abscalars struct {
	a float64
	b float64
}

//Settings contains all the paramters needed to build a batchnorm layer
type Settings struct {
	Mode    gocudnn.BatchNormMode `json:"mode,omitempty"`
	Managed bool                  `json:"managed,omitempty"`
}

//PerActivationPreset will presetup some values for the batch norm PerActivation
func PerActivationPreset(handle *cudnn.Handler, managed bool) (*Layer, error) {
	//	b, err := batchnorm.PreStagePerActivation(handle, managed)
	var flg gocudnn.BatchNormModeFlag
	fw := abscalars{
		a: alphaforwarddefault,
		b: betaforwarddefault,
	}
	bwd := abscalars{
		a: alphabackwarddefault,
		b: betabackwarddefault,
	}
	bwp := abscalars{
		a: alphabackwardparamdefault,
		b: betabackwardparamdefault,
	}
	return &Layer{

		fw:      fw,
		bwp:     bwp,
		bwd:     bwd,
		eps:     float64(2e-5),
		mode:    flg.PerActivation(),
		managed: managed,
	}, nil
}

//SpatialPreset will presetup some values for the batch norm Spatial Mode
func SpatialPreset(handle *cudnn.Handler, managed bool) (*Layer, error) {
	//	b, err := batchnorm.PreStageSpatial(handle, managed)
	var flg gocudnn.BatchNormModeFlag
	fw := abscalars{
		a: alphaforwarddefault,
		b: betaforwarddefault,
	}
	bwd := abscalars{
		a: alphabackwarddefault,
		b: betabackwarddefault,
	}
	bwp := abscalars{
		a: alphabackwardparamdefault,
		b: betabackwardparamdefault,
	}
	return &Layer{
		//b:    b,
		fw:      fw,
		bwp:     bwp,
		bwd:     bwd,
		eps:     float64(2e-5),
		mode:    flg.Spatial(),
		managed: managed,
	}, nil

}

//SpatialPersistantPreset will presetup some values for the batch norm SpatialPersistantPreset Mode
func SpatialPersistantPreset(handle *cudnn.Handler, managed bool) (*Layer, error) {
	//	b, err := batchnorm.PreStageSpatialPersistant(handle, managed)
	var flg gocudnn.BatchNormModeFlag
	fw := abscalars{
		a: alphaforwarddefault,
		b: betaforwarddefault,
	}
	bwd := abscalars{
		a: alphabackwarddefault,
		b: betabackwarddefault,
	}
	bwp := abscalars{
		a: alphabackwardparamdefault,
		b: betabackwardparamdefault,
	}
	return &Layer{
		//	b:    b,
		fw:      fw,
		bwp:     bwp,
		bwd:     bwd,
		eps:     float64(2e-5),
		mode:    flg.SpatialPersistent(),
		managed: managed,
	}, nil

}

//SetupPreset will allocate all the memory needed for the batch norm with the values passed when using one of the Preset functions
func (l *Layer) SetupPreset(handle *cudnn.Handler, x *layers.IO) error {
	var err error

	l.b, err = batchnorm.Stage(handle, x.T(), l.mode, l.managed)
	return err
}

//LayerSetup sets the layer up. I set the defaults for alpha and beta (a,b) for the forward(1,0), backward param(1,1), and backward data(1,0) that are used in cudnn.
//I am 70 percent sure that fwd and bwd data are set correctly.  I am about 25% sure bwd param is set correctly.  I will change it if it needs it
func LayerSetup(handle *cudnn.Handler, x *layers.IO, mode gocudnn.BatchNormMode, managed bool) (*Layer, error) {
	b, err := batchnorm.Stage(handle, x.T(), mode, managed)
	fw := abscalars{
		a: alphaforwarddefault,
		b: betaforwarddefault,
	}
	bwd := abscalars{
		a: alphabackwarddefault,
		b: betabackwarddefault,
	}
	bwp := abscalars{
		a: alphabackwardparamdefault,
		b: betabackwardparamdefault,
	}

	return &Layer{
		b:       b,
		fw:      fw,
		bwp:     bwp,
		bwd:     bwd,
		eps:     float64(2e-5),
		mode:    mode,
		managed: managed,
	}, err
}

//ForwardProp Does the Training Forward Prop of batch norm layer
func (l *Layer) ForwardProp(
	handle *cudnn.Handler,
	x,
	y *layers.IO,
) error {
	l.af = (1.0 / (1.0 + float64(l.counter)))

	err := l.b.ForwardTraining(handle, l.fw.a, l.fw.b, l.af, l.eps, x.T(), y.T())
	l.counter++
	return err
}

//BackProp does the back propagation in training the layer
func (l *Layer) BackProp(handle *cudnn.Handler, x, y *layers.IO) error {
	return l.b.BackwardProp(handle,
		l.bwp.a,
		l.bwp.b,
		l.bwd.a,
		l.bwd.b,
		l.eps,
		x.T(),
		x.DeltaT(),
		y.DeltaT())
}

//SetEps sets epsilon
func (l *Layer) SetEps(eps float64) {
	l.eps = eps
}

//SetBWPAlpha sets the alpha for bwards params
func (l *Layer) SetBWPAlpha(a float64) {
	l.bwp.a = a
}

//SetBWPBeta sets the beta for bwards params
func (l *Layer) SetBWPBeta(b float64) {
	l.bwp.b = b
}

//SetBWDAlpha sets the alpha for bwards data
func (l *Layer) SetBWDAlpha(a float64) {
	l.bwd.a = a
}

//SetBWDBeta sets beta for bwards data beta
func (l *Layer) SetBWDBeta(b float64) {
	l.bwd.b = b
}

//SetFWAlpha sets fwd alpha
func (l *Layer) SetFWAlpha(a float64) {
	l.fw.a = a
}

//SetFWBeta Sets FwdBeta
func (l *Layer) SetFWBeta(b float64) {
	l.fw.b = b
}
