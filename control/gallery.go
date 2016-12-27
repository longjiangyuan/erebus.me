package control

import (
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"

	"erebus.me/model"
	"erebus.me/view"
)

func init() {
	exif.RegisterParsers(mknote.All...)

	photo := Gallery{"html/photo/"}
	pjax.StripPrefix("/gallery/", &photo)
}

type Gallery struct {
	root string
}

func (photo *Gallery) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var target *os.File
	filename := path.Join(photo.root, r.URL.Path)

	if file, err := os.Open(filename); err == nil {
		target = file
	} else if os.IsNotExist(err) {
		view.NotFound(w, r)
		return
	} else {
		panic(err)
	}
	defer target.Close()

	info, err := target.Stat()
	if err != nil {
		panic(err)
	}

	switch r.Method {
	case "GET":
		if info.IsDir() {
			photo.Index(target, w, r)
		} else {
			photo.View(target, w, r)
		}
	case "POST":
		if info.IsDir() {
			photo.Post(w, r)
		} else {
			view.NotFound(w, r)
		}
	}
}

func (photo *Gallery) Post(w http.ResponseWriter, r *http.Request) {
	if model.GetLogin(r) == nil {
		panic("Not login")
	}

	mulitpart, err := r.MultipartReader()
	if err != nil {
		panic(err)
	}

	redirect := false

	for {
		part, err := mulitpart.NextPart()
		if err != nil {
			break
		}
		defer part.Close()

		filename := part.FileName()
		if filename == "" {
			if part.FormName() == "upload" {
				redirect = true
			}
			io.Copy(ioutil.Discard, part)
		} else {
			dstname := path.Join(photo.root, r.URL.Path, filename)
			dst, err := os.Create(dstname)
			if err != nil {
				panic(err)
			}
			defer dst.Close()

			if _, err = io.Copy(dst, part); err != nil {
				panic(err)
			}
			if _, err = dst.Seek(0, os.SEEK_SET); err != nil {
				panic(err)
			}

			//make thumbnail image
			/*
				thumbnailFile, err := os.Create(dstname + ".thumb")
				if err != nil {
					panic(err)
				}
				defer thumbnailFile.Close()
				photo.thumbnail(dst, thumbnailFile)
				log.Printf("/photo/%s/%s: uploaded", r.URL.Path, filename)
			*/

			funcNum := r.URL.Query().Get("CKEditorFuncNum")
			if redirect {
				if funcNum != "" {
					http.Redirect(w, r, "/photo/"+r.URL.Path+"?CKEditorFuncNum="+funcNum, http.StatusFound)
				} else {
					http.Redirect(w, r, "/photo/"+r.URL.Path, http.StatusFound)
				}
			} else if funcNum != "" {
				fmt.Fprintf(w, `<script type="text/javascript">window.parent.CKEDITOR.tools.callFunction("%s", "/photo/%s", "");</script>`, funcNum, filename)
			}
		}
	}
}

func (photo *Gallery) openFile(pathname string) (*os.File, os.FileInfo, error) {
	filename := path.Join(photo.root, pathname)

	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	return file, info, nil
}

func (photo *Gallery) View(file *os.File, w http.ResponseWriter, r *http.Request) {
	var page struct {
		Paths    []PathInfo
		Path     string
		Filename string
		Dirname  string
		Exif     *ExifFormatter
		Next     string
		Prev     string
	}

	page.Paths = splitPath(r.URL.Path)
	page.Dirname, page.Filename = path.Split(r.URL.Path)
	page.Path = r.URL.Path

	x, err := exif.Decode(file)
	if err == nil {
		page.Exif = NewExifFormatter(x)
	}

	dirname := path.Join(photo.root, page.Dirname)
	dir, err := os.Open(dirname)
	if err != nil {
		panic(err)
	}
	defer dir.Close()

	infos, err := dir.Readdir(-1)
	for i, info := range infos {
		if info.IsDir() {
			continue
		}
		name := info.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		if name == page.Filename {
			if i > 0 {
				page.Prev = infos[i-1].Name()
			}
			if i+1 < len(infos) {
				page.Next = infos[i+1].Name()
			}
			break
		}
	}
	view.Render(w, "/photo/view.html", &page)
}

type PathInfo struct {
	Name string
	Path string
}

func splitPath(pathname string) []PathInfo {
	paths := []PathInfo{}
	if pathname == "" {
		return paths
	}

	s := pathname
	if s[0] == '/' {
		s = s[1:]
	}
	for len(s) > 0 {
		var info PathInfo

		if idx := strings.IndexByte(s, '/'); idx >= 0 {
			info.Name = s[:idx]
			s = s[idx+1:]
		} else {
			info.Name = s
			s = ""
		}

		if len(paths) > 0 {
			info.Path = paths[len(paths)-1].Path + "/" + info.Name
		} else {
			info.Path = info.Name
		}
		paths = append(paths, info)
	}
	return paths
}

func (photo *Gallery) Index(dir *os.File, w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/") {
		r.URL.Path += "/"
	}

	if r.FormValue("CKEditorFuncNum") != "" {
		if model.GetLogin(r) == nil {
			view.Login(w, r)
			return
		}
	}

	var data struct {
		CKEditorFuncNum string
		Root            string
		Files           []string
		Dirs            []string
		Paths           []PathInfo
	}

	data.Paths = splitPath(r.URL.Path)
	data.Root = r.URL.Path
	data.CKEditorFuncNum = r.FormValue("CKEditorFuncNum")

	infos, err := dir.Readdir(-1)
	if err != nil {
		panic(err)
	}

	for _, info := range infos {
		filename := info.Name()
		//log.Printf("photo: index %s: name:%s dir:%v", root, filename, info.IsDir())

		if strings.HasPrefix(filename, ".") {
			continue
		}
		if strings.HasSuffix(filename, ".thumb") {
			continue
		}
		if info.IsDir() {
			data.Dirs = append(data.Dirs, filename)
		} else {
			data.Files = append(data.Files, filename)
		}
	}
	view.Render(w, "/photo/index.html", data)
}

/*
func (photo *Photo) readImageFile(filename string) (image.Image, string) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	image, format, err := image.Decode(file)
	if err != nil {
		panic(err)
	}
	return image, format
}


func (photo *Photo) thumbnail(src io.Reader, dst io.Writer) {
	//make thumbnail image
	img, _, err := image.Decode(src)
	if err != nil {
		panic(err)
	}
	thumbnail := photo.thumbnailImage(img)
	png.Encode(dst, thumbnail)
}

func (photo *Photo) thumbnailImage(src image.Image) *image.RGBA {
	r := src.Bounds()
	w := 200
	h := int(float64(r.Dy()) / float64(r.Dx()) * float64(w))

	if w == 0 || h == 0 || r.Dx() <= 0 || r.Dy() <= 0 {
		return image.NewRGBA(image.Rect(0, 0, w, h))
	}
	curw, curh := r.Dx(), r.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Get a source pixel.
			subx := x * curw / w
			suby := y * curh / h
			r32, g32, b32, a32 := src.At(subx, suby).RGBA()
			r := uint8(r32 >> 8)
			g := uint8(g32 >> 8)
			b := uint8(b32 >> 8)
			a := uint8(a32 >> 8)
			dst.SetRGBA(x, y, color.RGBA{r, g, b, a})
		}
	}
	return dst
}
*/
