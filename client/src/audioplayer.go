package client

/*
#cgo pkg-config: gstreamer-1.0
#include <gst/gst.h>
// ******************** 定义消息处理函数 ********************
gboolean bus_call(GstBus *bus, GstMessage *msg, gpointer data)
{
	GMainLoop *loop = (GMainLoop *)data;//这个是主循环的指针，在接受EOS消息时退出循环
	gchar *debug;
	GError *error;
	switch (GST_MESSAGE_TYPE(msg)) {
	case GST_MESSAGE_EOS:
		g_main_loop_quit(loop);
		//g_print("EOF\n");
		break;
	case GST_MESSAGE_ERROR:
		gst_message_parse_error(msg,&error,&debug);
		g_free(debug);
		g_printerr("ERROR:%s\n",error->message);
		g_error_free(error);
		g_main_loop_quit(loop);
		break;
	default:
		break;
	}
	return TRUE;
}
static GstBus *pipeline_get_bus(void *pipeline)
{
	return gst_pipeline_get_bus(GST_PIPELINE(pipeline));
}
static void bus_add_watch(void *bus, void *loop)
{
	gst_bus_add_watch(bus, bus_call, loop);
	gst_object_unref(bus);
}
static void set_path(void *play, gchar *path)
{
	g_object_set(G_OBJECT(play), "uri", path, NULL);
}
static void object_unref(void *pipeline)
{
	gst_object_unref(GST_OBJECT(pipeline));
}
static void media_ready(void *pipeline)
{
	gst_element_set_state(pipeline, GST_STATE_READY);
}
static void media_pause(void *pipeline)
{
	gst_element_set_state(pipeline, GST_STATE_PAUSED);
}
static void media_play(void *pipeline)
{
	gst_element_set_state(pipeline, GST_STATE_PLAYING);
}
static void media_stop(void *pipeline)
{
	gst_element_set_state(pipeline, GST_STATE_NULL);
}
static void set_mute(void *play)
{
	g_object_set(G_OBJECT(play), "mute", FALSE, NULL);
}
static void set_volume(void *play, int vol)
{
	int ret = vol % 101;
	g_object_set(G_OBJECT(play), "volume", ret/10.0, NULL);
}
static void media_seek(void *pipeline, gint64 pos)
{
	gint64 cpos;
	gst_element_query_position (pipeline, GST_FORMAT_TIME, &cpos);
	cpos += pos*1000*1000*1000;
	if (!gst_element_seek (pipeline, 1.0, GST_FORMAT_TIME, GST_SEEK_FLAG_FLUSH,
                         GST_SEEK_TYPE_SET, cpos,
                         GST_SEEK_TYPE_NONE, GST_CLOCK_TIME_NONE)) {
    		g_print ("Seek failed!\n");
    	}
}
*/
import "C"

import (
	"container/list"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"
	"unsafe"
)

type PlayStyle int

const (
	ORDER   PlayStyle = 0x100
	SINGLE  PlayStyle = 0x200
	SLOOP   PlayStyle = 0x300
	ALOOP   PlayStyle = 0x400
	SHUFFLE PlayStyle = 0x500
)

type AudioPlayer struct {
	aList        *list.List // play list
	volumeSize   int        // audio volume size
	playStyle    PlayStyle  // play order style
	isOutOfOrder bool       // is the file list is already out of order
	flag         chan byte  // a flag to indicate pause, play, next, etc
	seekLen      int        // seek length for second
	isPausing    bool
	isPlaying    bool
}

//var g_list *list.List
//var g_wg *sync.WaitGroup
//var g_isQuit bool = false
//var g_play_style int
//var g_isOutOfOrder bool
//var g_volume_size int = 10

func gString(s string) *C.gchar {
	return (*C.gchar)(C.CString(s))
}

func gFree(s unsafe.Pointer) {
	C.g_free(C.gpointer(s))
}

func NewAudioPlayer() AudioPlayer {
	var a AudioPlayer
	a.aList = list.New()
	a.volumeSize = 10
	a.playStyle = ORDER
	a.flag = make(chan byte)
	return a
}

func (a *AudioPlayer) walkFunc(fPath string, info os.FileInfo, err error) error {
	if info.IsDir() {
		return nil
	}
	switch filepath.Ext(fPath) {
	case ".mp3":
	case ".wav":
	case ".ogg":
	case ".wma":
	case ".rmvb":
	default:
		return nil
	}
	if x, err0 := filepath.Abs(fPath); err != nil {
		err = err0
		return err
	} else {
		p := fmt.Sprintf("file://%s", x)
		a.aList.PushBack(p)
	}

	return err
}

// make file out of order
func (a *AudioPlayer) outOfOrder() {
	iTotal := 25
	if iTotal > a.aList.Len() {
		iTotal = a.aList.Len()
	}
	ll := make([]*list.List, iTotal)

	for i := 0; i < iTotal; i++ {
		ll[i] = list.New()
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for e := a.aList.Front(); e != nil; e = e.Next() {
		fPath, ok := e.Value.(string)
		if !ok {
			panic("The path is invalid string")
		}
		if rand.Int()%2 == 0 {
			ll[r.Intn(iTotal)].PushFront(fPath)
		} else {
			ll[r.Intn(iTotal)].PushBack(fPath)
		}
	}

	r0 := rand.New(rand.NewSource(time.Now().UnixNano()))
	a.aList.Init()
	for i := 0; i < iTotal; i++ {
		if r0.Intn(2) == 0 {
			a.aList.PushBackList(ll[i])
		} else {
			a.aList.PushFrontList(ll[i])
		}
		ll[i].Init()
	}
}

func (a *AudioPlayer) playProcess() {
	var pipeline *C.GstElement // 定义组件
	var bus *C.GstBus
	var g_wg sync.WaitGroup
	var loop *C.GMainLoop
	wg := new(sync.WaitGroup)
	C.gst_init((*C.int)(unsafe.Pointer(nil)),
		(***C.char)(unsafe.Pointer(nil)))
	loop = C.g_main_loop_new((*C.GMainContext)(unsafe.Pointer(nil)),
		C.gboolean(0)) // 创建主循环，在执行 g_main_loop_run后正式开始循环

	sig_out := make(chan bool)

	g_wg.Add(1)
	defer close(sig_out)
	defer g_wg.Done()
	if a.isOutOfOrder {
		a.outOfOrder()
		debug.FreeOSMemory()
	}

	start := a.aList.Front()
	end := a.aList.Back()
	e := a.aList.Front()

	v0 := gString("playbin")
	v1 := gString("play")
	pipeline = C.gst_element_factory_make(v0, v1)
	gFree(unsafe.Pointer(v0))
	gFree(unsafe.Pointer(v1))
	// 得到 管道的消息总线
	bus = C.pipeline_get_bus(unsafe.Pointer(pipeline))
	if bus == (*C.GstBus)(nil) {
		fmt.Println("GstBus element could not be created.Exiting.")
		return
	}
	C.bus_add_watch(unsafe.Pointer(bus), unsafe.Pointer(loop))
	// 开始循环

	g_isQuit := false

	go func(sig_quit chan bool) {
		wg.Add(1)
		i := 0
	LOOP_RUN:
		for !g_isQuit {
			if i != 0 {
				C.media_ready(unsafe.Pointer(pipeline))
				C.media_play(unsafe.Pointer(pipeline))
			}
			C.g_main_loop_run(loop)
			C.media_stop(unsafe.Pointer(pipeline))
			switch a.playStyle {
			case SINGLE:
				sig_quit <- true
				break LOOP_RUN

			case ORDER:
				if e != end {
					e = e.Next()
				} else {
					break LOOP_RUN
				}
			case SHUFFLE:
				if e != end {
					e = e.Next()
				} else {
					break LOOP_RUN
				}

			case SLOOP:

			case ALOOP:
				if e != end {
					e = e.Next()
				} else {
					e = start
				}

			}
			fPath, ok := e.Value.(string)
			if ok {
				v2 := gString(fPath)
				C.set_path(unsafe.Pointer(pipeline), v2)
				gFree(unsafe.Pointer(v2))

			} else {
				break
			}
			i++
		}

		C.object_unref(unsafe.Pointer(pipeline))
		wg.Done()

	}(sig_out)

	fPath, ok := e.Value.(string)
	if ok {
		v2 := gString(fPath)
		C.set_path(unsafe.Pointer(pipeline), v2)
		gFree(unsafe.Pointer(v2))

		C.media_ready(unsafe.Pointer(pipeline))
		C.media_play(unsafe.Pointer(pipeline))
		//C.set_mute(unsafe.Pointer(pipeline))

		lb := true
		for lb {
			select {
			case op := <-a.flag:
				switch op {
				case 's':
					C.media_pause(unsafe.Pointer(pipeline))
				case 'r':
					C.media_play(unsafe.Pointer(pipeline))
				case 'n':
					switch a.playStyle {
					case SINGLE:
						lb = false
						g_isQuit = true
					case ORDER:
						fallthrough
					case SHUFFLE:
						C.media_stop(unsafe.Pointer(pipeline))
						if e != end {
							e = e.Next()
						} else {
							lb = false
							g_isQuit = true
						}
					case SLOOP:
						C.media_stop(unsafe.Pointer(pipeline))

					case ALOOP:
						if e != end {
							e = e.Next()
						} else {
							e = start
						}

					}
					if !lb {
						fPath, ok := e.Value.(string)
						if ok {
							v2 := gString(fPath)
							C.set_path(unsafe.Pointer(pipeline), v2)
							gFree(unsafe.Pointer(v2))
							C.media_ready(unsafe.Pointer(pipeline))
							C.media_play(unsafe.Pointer(pipeline))
						} else {
							lb = false
							g_isQuit = true
						}
					}
				//C.g_main_loop_quit(loop)
				case 'p':
					switch a.playStyle {
					case SINGLE:
					// do nothing ???
					case ORDER:
						fallthrough
					case SHUFFLE:

						C.media_stop(unsafe.Pointer(pipeline))
						if e != start {
							e = e.Prev()
							fPath, ok := e.Value.(string)
							if ok {
								v2 := gString(fPath)
								C.set_path(unsafe.Pointer(pipeline), v2)
								gFree(unsafe.Pointer(v2))
								C.media_ready(unsafe.Pointer(pipeline))
								C.media_play(unsafe.Pointer(pipeline))
							} else {
								lb = false
								g_isQuit = true
							}
						} else {
							lb = false
							g_isQuit = true
						}
					case SLOOP:
						C.media_stop(unsafe.Pointer(pipeline))
						fpath, ok := e.Value.(string)
						if ok {
							v2 := gString(fpath)
							C.set_path(unsafe.Pointer(pipeline), v2)
							gFree(unsafe.Pointer(v2))
							C.media_ready(unsafe.Pointer(pipeline))
							C.media_play(unsafe.Pointer(pipeline))
						}
					case ALOOP:
						C.media_stop(unsafe.Pointer(pipeline))
						if e != start {
							e = e.Prev()
						} else {
							e = end
						}
						fPath, ok := e.Value.(string)
						if ok {
							v2 := gString(fPath)
							C.set_path(unsafe.Pointer(pipeline), v2)
							gFree(unsafe.Pointer(v2))
							C.media_ready(unsafe.Pointer(pipeline))
							C.media_play(unsafe.Pointer(pipeline))
						}
					}

				case 'q':
					lb = false
					g_isQuit = true
				case '+':
					a.volumeSize++
					C.set_volume(unsafe.Pointer(pipeline), C.int(a.volumeSize))
				case '-':
					a.volumeSize--
					if a.volumeSize < 0 {
						a.volumeSize = 0
					}
					C.set_volume(unsafe.Pointer(pipeline), C.int(a.volumeSize))
				case 't':
					C.media_seek(unsafe.Pointer(pipeline), C.gint64(5))

				}
			case vv0 := <-sig_out:
				if vv0 {
					C.g_main_loop_quit(loop)
					wg.Wait()
					g_wg.Done()
					g_wg.Wait()
					close(sig_out)
					os.Exit(0)
				}
			}
		}

	} else {
		// 路径非法
		return
	}

	C.g_main_loop_quit(loop)
	wg.Wait()
}

func (a *AudioPlayer) SetPlayStyle(style PlayStyle) {
	a.playStyle = style
}

func (a *AudioPlayer) AddFile(mfile string) error {
	p, err := filepath.Abs(mfile)
	if err != nil {
		return fmt.Errorf("Error: %v\n", err)
	}
	mfile = fmt.Sprintf("file://%s", p)
	a.aList.PushBack(mfile)
	return nil
}

func (a *AudioPlayer) SetList(playlist *list.List) {
	if playlist != nil {
		a.aList = playlist
	}
}

func (a *AudioPlayer) AddDir(mdir string) error {
	return filepath.Walk(mdir, a.walkFunc)
}

func (a *AudioPlayer) Play() {
	// when the play is already start but paused, resume it but not play new one
	if (a.isPausing) {
		a.Resume()
	} else {
		go a.playProcess()
		a.isPlaying = true
	}
}

func (a *AudioPlayer) Pause() {
	a.flag <- 's'
	a.isPausing = true
}

func (a *AudioPlayer) Seek(len int) {
	a.seekLen = len
	a.flag <- 't'
}

func (a *AudioPlayer) Quit() {
	a.flag <- 'q'
	a.isPlaying = false
}

func (a *AudioPlayer) Next() {
	a.flag <- 'n'
}

func (a *AudioPlayer) Previous() {
	a.flag <- 'p'
}

func (a *AudioPlayer) Resume() {
	if (a.isPausing) {
		a.flag <- 'r'
		a.isPausing = false
	}
}

func (a *AudioPlayer) IncreaseVolume() {
	a.flag <- '+'
}

func (a *AudioPlayer) DecreaseVolume() {
	a.flag <- '-'
}

func (a *AudioPlayer) IsPlaying() bool{
	return a.isPlaying&&(!a.isPausing)
}

func (a *AudioPlayer) IsPausing() bool{
	return a.isPausing
}
