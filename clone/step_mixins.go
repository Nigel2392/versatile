package clone

import (
	"context"
	"iter"
	"reflect"

	"github.com/Nigel2392/versatile/mixins"
)

type (
	PreCopyMixinDefiner = interface {
		Step
		PreCopyMixins() []PreCopyMixin
	}

	PostCopyMixinDefiner = interface {
		Step
		PostCopyMixins() []PostCopyMixin
	}
)

type PreCopyMixin interface {
	Step
	PreCopy(ctx context.Context, state *State, dst, src reflect.Value) error
}

type PostCopyMixin interface {
	Step
	PostCopy(ctx context.Context, state *State, dst, src reflect.Value) error
}

//	func copyMixinsFunc[IN, OUT any](obj Step, call func(IN) []OUT) iter.Seq2[OUT, int] {
//		return mixins.MixinsFunc(obj, false, func(obj Step, depth int) iter.Seq[OUT] {
//			d, ok := obj.(IN)
//			if !ok {
//				return noYield[func(OUT) bool]
//			}
//
//			return func(yield func(OUT) bool) {
//				for _, o := range call(d) {
//					if !yield(o) {
//						break
//					}
//				}
//			}
//		})
//	}
//
//	func PreCopyMixins(obj Step) iter.Seq2[PreCopyMixin, int] {
//		return copyMixinsFunc(obj, PreCopyMixinDefiner.PreCopyMixins)
//	}
//
//	func PostCopyMixins(obj Step) iter.Seq2[PostCopyMixin, int] {
//		return copyMixinsFunc(obj, PostCopyMixinDefiner.PostCopyMixins)
//	}

func noYield[T any](_ T) {}

func PreCopyMixins(obj Step) iter.Seq2[PreCopyMixin, int] {
	return mixins.MixinsFunc(obj, false, func(obj Step, depth int) iter.Seq[PreCopyMixin] {
		d, ok := obj.(PreCopyMixinDefiner)
		if !ok {
			return noYield[func(PreCopyMixin) bool]
		}

		return func(yield func(PreCopyMixin) bool) {
			for _, o := range d.PreCopyMixins() {
				if !yield(o) {
					break
				}
			}
		}
	})
}

func PostCopyMixins(obj Step) iter.Seq2[PostCopyMixin, int] {
	return mixins.MixinsFunc(obj, false, func(obj Step, depth int) iter.Seq[PostCopyMixin] {
		d, ok := obj.(PostCopyMixinDefiner)
		if !ok {
			return noYield[func(PostCopyMixin) bool]
		}

		return func(yield func(PostCopyMixin) bool) {
			for _, o := range d.PostCopyMixins() {
				if !yield(o) {
					break
				}
			}
		}
	})
}
