package activation

import (
	"errors"

	"github.com/dereklstinson/GoCuNets/cudnn"
	"github.com/dereklstinson/GoCuNets/cudnn/tensor"
	"github.com/dereklstinson/GoCudnn"
)

//Ops is the non linear function that is used in neural networks. This structure holds the information used to performing the activation function.
type Ops struct {
	helper gocudnn.Activation
	desc   *gocudnn.ActivationD
	xdesc  *gocudnn.XActivationD
	mode   Mode
	nan    cudnn.NanMode
}

const defaultparachantrainingmode = gocudnn.TrainingMode(4) //This is adam
const defaultcoefforleaky = .01
const defaultcoefforclipped = 6

//Stage creates an activation struct given the properties passed in function
func Stage(handle cudnn.Handler, mode Mode, tmode cudnn.TrainMode, nan cudnn.NanMode, coef float64) (*Ops, error) {
	var dtype gocudnn.DataTypeFlag
	var xtra gocudnn.Xtra
	var mflg ModeFlag
	var hlp gocudnn.Activation
	switch mode {
	case mflg.AdvancedThreshRandRelu():
		desc, err := xtra.NewXActivationDescriptor(handle.XHandle(), mode.x(), tmode.Cu(), dtype.Float(), coef)
		if err != nil {
			return nil, err
		}
		return &Ops{
			xdesc: desc,
			mode:  mode,
		}, err
	case mflg.Leaky():
		desc, err := xtra.NewXActivationDescriptor(handle.XHandle(), mode.x(), tmode.Cu(), dtype.Float(), coef)
		if err != nil {
			return nil, err
		}
		return &Ops{
			xdesc: desc,
			mode:  mode,
		}, err
	case mflg.PRelu():
		desc, err := xtra.NewXActivationDescriptor(handle.XHandle(), mode.x(), tmode.Cu(), dtype.Float(), coef)
		if err != nil {
			return nil, err
		}
		return &Ops{
			xdesc: desc,
			mode:  mode,
		}, err
	case mflg.ClippedRelu():
		x, err := hlp.NewActivationDescriptor(mode.c(), nan.Cu(), coef)
		return &Ops{
			desc: x,
			mode: mode,
		}, err
	case mflg.Elu():
		x, err := hlp.NewActivationDescriptor(mode.c(), nan.Cu(), coef)
		return &Ops{
			desc: x,
			mode: mode,
		}, err

	case mflg.Relu():
		x, err := hlp.NewActivationDescriptor(mode.c(), nan.Cu(), coef)
		return &Ops{
			desc: x,
			mode: mode,
		}, err
	case mflg.Sigmoid():
		x, err := hlp.NewActivationDescriptor(mode.c(), nan.Cu(), coef)
		return &Ops{
			desc: x,
			mode: mode,
		}, err
	case mflg.Tanh():
		x, err := hlp.NewActivationDescriptor(mode.c(), nan.Cu(), coef)
		return &Ops{
			desc: x,
			mode: mode,
		}, err

	}
	return nil, errors.New("Not supported activation")
}

/*
switch Mode{
case mflg.AdvancedThreshRandRelu():
case mflg.ClippedRelu():
case mflg.Elu():
case mflg.Leaky():
case mflg.PRelu():
case mflg.Relu():
case mflg.Sigmoid():
case mflg.Tanh():

}
*/
/*
//ReStage will destroy the desc in the Op and then make a new one to the settings given.
func (act *Ops) ReStage(mode gocudnn.ActivationMode, nan gocudnn.PropagationNAN, coef float64) error {
	err := act.desc.DestroyDescriptor()
	if err != nil {
		return err
	}
	act.desc, err =.NewActivationDescriptor(mode, nan, coef)
	return err
}
*/

//Info returns the Info struct.  (Made for saving to a json file at a higher level)
func (act *Ops) Info() (OpInfo, error) {

	amode, propnan, coef, err := act.Properties()

	return OpInfo{
		Mode:    Mode(amode),
		NanProp: cudnn.NanMode(propnan),
		Coef:    coef,
	}, err
}

//Properties returns the values that were used to Create the Activation struct
func (act *Ops) Properties() (Mode, cudnn.NanMode, float64, error) {
	a, b, c, err := act.desc.GetDescriptor()
	return Mode(a), cudnn.NanMode(b), c, err

}

//FwdProp is the forward propigation function for the Activation struct
func (act *Ops) FwdProp(
	handle *cudnn.Handler,
	alpha float64,
	x *tensor.Volume,
	beta float64,
	y *tensor.Volume,
	alphas *tensor.Volume,
	betas *tensor.Volume) error {
	_, dtypex, _, err := x.Properties()
	if err != nil {
		return err
	}
	_, dtypey, _, err := y.Properties()

	if err != nil {
		return err
	}
	if dtypex != dtypey {
		return errors.New("output type not matching input type")
	}
	a := gocudnn.CScalarByDataType(dtypex, alpha)
	b := gocudnn.CScalarByDataType(dtypex, beta)
	if a == nil || b == nil {
		return errors.New("Unsupported Datatype for either alpha or beta")
	}

	return act.desc.Forward(handle.Cudnn(), a, x.TD(), x.Memer(), b, y.TD(), y.Memer())
}

//BwdProp is the backwards propigation of the activation struct
func (act *Ops) BwdProp(
	handle *cudnn.Handler,
	alpha float64,
	y *tensor.Volume,
	dy *tensor.Volume,
	x *tensor.Volume,
	beta float64,
	dx *tensor.Volume,
) error {
	_, dtypedx, _, err := dx.Properties()
	if err != nil {
		return err
	}
	_, dtypex, _, err := x.Properties()
	if err != nil {
		return err
	}
	_, dtypey, _, err := y.Properties()
	if err != nil {
		return err
	}
	_, dtypedy, _, err := dy.Properties()
	if err != nil {
		return err
	}
	if dtypedx != dtypey || dtypedx != dtypedy || dtypedx != dtypex {
		return errors.New("output type not matching input type")
	}
	a := gocudnn.CScalarByDataType(dtypedx, alpha)
	b := gocudnn.CScalarByDataType(dtypedx, beta)
	if a == nil || b == nil {
		return errors.New("Unsupported Datatype for either alpha or beta")
	}

	return act.desc.Backward(handle.Cudnn(), a, y.TD(), y.Memer(), dy.TD(), dy.Memer(), x.TD(), x.Memer(), b, dx.TD(), dx.Memer())
}

//Destroy destroys the cuda allocated memory associated with Activation
func (act *Ops) Destroy() error {

	return act.desc.DestroyDescriptor()
}
