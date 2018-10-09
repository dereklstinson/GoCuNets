package cpu

import "errors"

//SegmentBatch1CHWtoNCHW4d Takes a Volume and Segments it into Batches to the size h,w given. and rounds up by one.  Values not used in new tensor will be zero
func SegmentBatch1CHWtoNCHW4d(values []float32, dims []int32, h, w int32) ([]float32, []int32, error) {
	if len(dims) != 4 {
		return nil, nil, errors.New("The Length of dims should equal 4")
	}
	if dims[0] != int32(1) {
		return nil, nil, errors.New("N value needs to be 1")
	}
	n1 := intceiling(dims[2], h)
	n2 := intceiling(dims[3], h)
	oHH := dims[2]
	oHW := dims[3]
	n := n1 * n2
	c := dims[1]
	newdims := []int32{n, c, h, w}
	//	totalvol := Volume(dims)
	v := make([]float32, n*c*h*w)
	striderh := int32(0)
	for i := int32(0); i < n1; i++ {
		striderw := int32(0)
		for j := int32(0); j < n2; j++ {
			for k := int32(0); k < c; k++ {
				for l := int32(0); l < h; l++ {
					oh := striderh + l
					for m := int32(0); m < w; m++ {
						ow := striderw + m
						if oh < oHH && ow < oHW {
							v[(i*n2*c*h*w)+(j*c*h*w)+(k*h*w)+(l*h)+m] = values[(k*oHW*oHH)+(oh*oHW)+(ow)]
						} else {
							v[(i*n2*c*h*w)+(j*c*h*w)+(k*h*w)+(l*h)+m] = 0
						}

					}
				}
			}
			striderw += w
		}
		striderh += h
	}
	return v, newdims, nil
}

func Volume(dims []int32) int32 {
	vol := int32(1)
	for i := 0; i < len(dims); i++ {
		vol *= dims[i]
	}
	return vol
}
func intceiling(a, b int32) int32 {
	return ((a - int32(1)) / b) + int32(1)
}
