// Copyright ©2012 The bíogo.cluster Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package meanshift provides mean shift clustering for ℝⁿ data.
package meanshift

import (
	"code.google.com/p/biogo.cluster"
	"fmt"
)

type pnt []float64

func (p pnt) V() []float64 { return p }

type value struct {
	pnt
	w       float64
	cluster int
}

func (v *value) Weight() float64 { return v.w }
func (v *value) Cluster() int    { return v.cluster }

type center struct {
	pnt
	w       float64
	indices cluster.Indices
}

func (c *center) Cluster() cluster.Indices { return c.indices }

type Shifter interface {
	Init(cluster.Interface)
	Shift() float64
	Bandwidth() float64
	Centers() ([]cluster.Center, []cluster.Indices)
}

// A MeanShift clusters ℝⁿ data according to the mean shift algorithm.
type MeanShift struct {
	k       Shifter
	tol     float64
	maxIter int
	values  []value
	centers []center
	ci      []cluster.Indices
}

// New creates a new mean shift Clusterer object populated with data from an Interface value, data
// and using the Kernel k.
func New(data cluster.Interface, k Shifter, tol float64, maxIter int) *MeanShift {
	k.Init(data)
	return &MeanShift{
		k:       k,
		tol:     tol,
		maxIter: maxIter,
		values:  convert(data),
	}
}

// Convert the data to the internal float64 representation.
func convert(data cluster.Interface) []value {
	va := make([]value, data.Len())
	for i := 0; i < data.Len(); i++ {
		va[i] = value{pnt: append(pnt(nil), data.Values(i)...)}
	}
	if w, ok := data.(cluster.Weighter); ok {
		for i := 0; i < data.Len(); i++ {
			va[i].w = w.Weight(i)
		}
	} else {
		for i := 0; i < data.Len(); i++ {
			va[i].w = 1
		}
	}

	return va
}

// Cluster the data using the mean shift algorithm.
func (ms *MeanShift) Cluster() error {
	for i := 0; ; i++ {
		delta := ms.k.Shift()
		if delta <= ms.tol {
			break
		}
		if i > ms.maxIter {
			return fmt.Errorf("meanshift: exceeded max iterations: delta=%f", delta)
		}
	}

	var cen []cluster.Center
	cen, ms.ci = ms.k.Centers()
	ms.centers = make([]center, len(cen))
	for i, c := range cen {
		ms.centers[i] = center{pnt: c.V(), indices: ms.ci[i]}
		for _, j := range ms.ci[i] {
			ms.values[j].cluster = i
		}
	}

	return nil
}

// Total calculates the total sum of squares for the data relative to the data mean.
func (ms *MeanShift) Total() float64 {
	p := make([]float64, len(ms.values[0].pnt))

	for _, v := range ms.values {
		for i := range p {
			p[i] += v.pnt[i]
		}
	}
	inv := 1 / float64(len(ms.values))
	for i := range p {
		p[i] *= inv
	}

	var ss float64
	for _, v := range ms.values {
		for i := range p {
			d := p[i] - v.pnt[i]
			ss += d * d
		}
	}

	return ss
}

// Within calculates the sum of squares within each cluster. It returns nil if Cluster
// has not been called.
func (ms *MeanShift) Within() []float64 {
	if ms.centers == nil {
		return nil
	}
	ss := make([]float64, len(ms.centers))

	for _, v := range ms.values {
		for i := range ms.centers[0].pnt {
			d := ms.centers[v.cluster].pnt[i] - v.pnt[i]
			ss[v.cluster] += d * d
		}
	}

	return ss
}

// Centers returns the cluster centers.
func (ms *MeanShift) Centers() []cluster.Center {
	cs := make([]cluster.Center, len(ms.centers))
	for i := range ms.centers {
		cs[i] = &ms.centers[i]
	}
	return cs
}

// Values returns returns a slice of the original data values in the MeanShift.
func (ms *MeanShift) Values() []cluster.Value {
	vs := make([]cluster.Value, len(ms.values))
	for i := range ms.values {
		vs[i] = &ms.values[i]
	}
	return vs
}