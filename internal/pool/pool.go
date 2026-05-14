package pool

import "github.com/panjf2000/ants/v2"

var Global *ants.Pool

func Init(size int) error {
	p, err := ants.NewPool(size)
	if err != nil {
		return err
	}
	Global = p
	return nil
}
func Submit(task func()) error {
	if Global == nil {
		if err := Init(100); err != nil {
			return err
		}
	}
	return Global.Submit(task)
}
func Release() {
	if Global != nil {
		Global.Release()
	}
}
