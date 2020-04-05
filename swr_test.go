package gmf

import (
	"log"
	"testing"
)

func TestSwrInit(t *testing.T) {
	options := []*Option{
		{"in_channel_count", 2},
		{"in_sample_rate", 44100},
		{"in_sample_fmt", AV_SAMPLE_FMT_S16},
		{"out_channel_count", 2},
		{"out_sample_rate", 44100},
		{"out_sample_fmt", AV_SAMPLE_FMT_S16},
	}

	swrCtx ,err := NewSwrCtx(options, 2,AV_SAMPLE_FMT_S16)
	if err != nil {
		t.Fatalf("error creating context %v\n" , err)
	}
	if swrCtx == nil {
		t.Fatal("unable to create Swr Context")
	} else {
		swrCtx.Free()
	}

	log.Println("Swr context is createad")
}

func TestSwrInitWithParams(t *testing.T) {
	swrCtx ,err := NewSwrCtxWithParams(2 , AV_SAMPLE_FMT_S16 , 48000 , 2 , AV_SAMPLE_FMT_S16 , 16000)
	if err != nil {
		t.Fatalf("error creating context %v\n" , err)
	}
	if swrCtx == nil {
		t.Fatal("unable to create Swr Context")
	} else {
		swrCtx.Free()
	}

	log.Println("Swr context is created with params")

}

func TestSwrConvertSameParams(t *testing.T){
	swrCtx ,err := NewSwrCtxWithParams(2 , AV_SAMPLE_FMT_S16 , 48000 , 2 , AV_SAMPLE_FMT_S16 , 48000)
	if err != nil {
		t.Fatalf("error creating context %v\n" , err)
	}

	defer swrCtx.Free()

	srcFrame,err := NewAudioFrame(AV_SAMPLE_FMT_S16,2,1024 )
	defer  srcFrame.Free()

	if err != nil {
		t.Fatalf("unable to create frame %v\n" , err)
	}

	dstFrame,err := swrCtx.Convert(srcFrame)
	if err != nil {
		t.Fatalf("unable to convert frame %v\n" , err)
	}

	defer dstFrame.Free()

	if srcFrame.Channels() != dstFrame.Channels() {
		t.Fatal("channel count mismatch")
	}

	if srcFrame.Format() != dstFrame.Format() {
		t.Fatal("sample format mismatch")
	}

	if srcFrame.NbSamples() != dstFrame.NbSamples() {
		t.Fatal("nb samples mismatch")
	}


	log.Println("all frame params are equal")

}

func TestSwrConvertDownSample(t *testing.T){
	swrCtx ,err := NewSwrCtxWithParams(2 , AV_SAMPLE_FMT_S16 , 48000 , 2 , AV_SAMPLE_FMT_S16 , 16000)
	if err != nil {
		t.Fatalf("error creating context %v\n" , err)
	}

	defer swrCtx.Free()

	srcFrame,err := NewAudioFrame(AV_SAMPLE_FMT_S16,2,1200 )
	defer  srcFrame.Free()

	if err != nil {
		t.Fatalf("unable to create frame %v\n" , err)
	}

	dstFrame,err := swrCtx.Convert(srcFrame)
	if err != nil {
		t.Fatalf("unable to convert frame %v\n" , err)
	}

	defer dstFrame.Free()

	if srcFrame.Channels() != dstFrame.Channels() {
		t.Fatal("channel count mismatch")
	}

	if srcFrame.Format() != dstFrame.Format() {
		t.Fatal("sample format mismatch")
	}

	if srcFrame.NbSamples() == dstFrame.NbSamples() {
		t.Fatal("nb samples should not match")
	}

	log.Printf("src nb samples %d , dst nb samples %d\n" , srcFrame.NbSamples() , dstFrame.NbSamples())


}
