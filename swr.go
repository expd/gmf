package gmf

/*

#cgo pkg-config: libswresample

#include "libswresample/swresample.h"
#include <libavcodec/avcodec.h>
#include <libavutil/frame.h>

int gmf_sw_resample(SwrContext* ctx, AVFrame* dstFrame, AVFrame* srcFrame){
	return swr_convert(ctx, dstFrame->data, dstFrame->nb_samples,
		(const uint8_t **)srcFrame->data, srcFrame->nb_samples);
}

int gmf_swr_flush(SwrContext* ctx, AVFrame* dstFrame) {
	return swr_convert(ctx, dstFrame->data, dstFrame->nb_samples,
		NULL, 0);
}

*/
import "C"

import (
	"fmt"
	"unsafe"
)

type SwrCtx struct {
	swrCtx       *C.struct_SwrContext
	channels     int
	format       int32
	sampleRate   int
	inSampleRate int
	layout       int
}

var AV_ROUND_UP uint32 = C.AV_ROUND_UP

//
func NewSwrCtxWithParams(in_channel_count int, in_sample_format int32, in_sample_rate int,
	out_channel_count int, out_sample_format int32, out_sample_rate int) (*SwrCtx, error) {
	in_channel_layout := C.int64_t(C.av_get_default_channel_layout(C.int(in_channel_count)))
	out_channel_layout := C.int64_t(C.av_get_default_channel_layout(C.int(out_channel_count)))

	var swrCtx *C.struct_SwrContext = nil

	swrCtx = C.swr_alloc_set_opts(swrCtx,
		out_channel_layout,
		out_sample_format,
		C.int(out_sample_rate),
		in_channel_layout,
		in_sample_format,
		C.int(in_sample_rate),
		0,
		unsafe.Pointer(nil))

	ctx := &SwrCtx{
		swrCtx:       swrCtx,
		channels:     out_channel_count,
		format:       out_sample_format,
		sampleRate:   out_sample_rate,
		inSampleRate: in_sample_rate,
		layout:       int(out_channel_layout),
	}

	if ret := int(C.swr_init(ctx.swrCtx)); ret < 0 {
		return nil, fmt.Errorf("error initializing swr context - %s", AvError(ret))
	}

	return ctx, nil

	// init with sws

}

func NewSwrCtx(options []*Option, channels int, format int32) (*SwrCtx, error) {
	ctx := &SwrCtx{
		swrCtx:   C.swr_alloc(),
		channels: channels,
		format:   format,
	}

	for _, option := range options {
		option.Set(ctx.swrCtx)
	}

	if ret := int(C.swr_init(ctx.swrCtx)); ret < 0 {
		return nil, fmt.Errorf("error initializing swr context - %s", AvError(ret))
	}

	return ctx, nil
}

func (ctx *SwrCtx) Free() {
	C.swr_free(&ctx.swrCtx)
}

func (ctx *SwrCtx) Convert(input *Frame) (*Frame, error) {
	var (
		dst *Frame
		err error
	)

	out_samples := int(C.av_rescale_rnd(
		C.swr_get_delay(ctx.swrCtx, C.int64_t(ctx.inSampleRate))+
			C.int64_t(input.NbSamples()),
		C.int64_t(ctx.sampleRate),
		C.int64_t(ctx.inSampleRate),
		AV_ROUND_UP))

	if dst, err = NewAudioFrame(ctx.format, ctx.channels, out_samples); err != nil {
		return nil, fmt.Errorf("error creating new audio frame - %s\n", err)
	}

	dst.SetChannelLayout(ctx.layout)

	C.gmf_sw_resample(ctx.swrCtx, dst.avFrame, input.avFrame)

	return dst, nil
}

func (ctx *SwrCtx) Flush(nbSamples int) (*Frame, error) {
	var (
		dst *Frame
		err error
	)

	if dst, err = NewAudioFrame(ctx.format, ctx.channels, nbSamples); err != nil {
		return nil, fmt.Errorf("error creating new audio frame - %s\n", err)
	}
	dst.SetChannelLayout(ctx.layout)
	C.gmf_swr_flush(ctx.swrCtx, dst.avFrame)

	return dst, nil
}

func DefaultResampler(ost *Stream, frames []*Frame, flush bool) []*Frame {
	var (
		result             []*Frame = make([]*Frame, 0)
		winFrame, tmpFrame *Frame
	)

	if ost.SwrCtx == nil || ost.AvFifo == nil {
		return frames
	}

	frameSize := ost.CodecCtx().FrameSize()

	for i, _ := range frames {
		ost.AvFifo.Write(frames[i])

		for ost.AvFifo.SamplesToRead() >= frameSize {
			winFrame = ost.AvFifo.Read(frameSize)
			winFrame.SetChannelLayout(ost.CodecCtx().GetDefaultChannelLayout(ost.CodecCtx().Channels()))

			tmpFrame, _ = ost.SwrCtx.Convert(winFrame)
			if tmpFrame == nil || tmpFrame.IsNil() {
				break
			}

			tmpFrame.SetPts(ost.Pts)
			tmpFrame.SetPktDts(int(ost.Pts))

			ost.Pts += int64(frameSize)

			result = append(result, tmpFrame)
		}
	}

	if flush {
		if tmpFrame, _ = ost.SwrCtx.Flush(frameSize); tmpFrame != nil && !tmpFrame.IsNil() {
			tmpFrame.SetPts(ost.Pts)
			tmpFrame.SetPktDts(int(ost.Pts))

			ost.Pts += int64(frameSize)

			result = append(result, tmpFrame)
		}
	}

	for i := 0; i < len(frames); i++ {
		if frames[i] != nil {
			frames[i].Free()
		}
	}

	return result
}
