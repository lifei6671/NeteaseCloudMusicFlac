package main

import (
	"fmt"
	"os"
	"strings"
	"net/http"
	"net/url"
	"io/ioutil"
	"regexp"
	"encoding/json"
	"sync"
	"io"
)

const(
	SuggestionUrl = "http://sug.music.baidu.com/info/suggestion";
	Fmlink = "http://music.baidu.com/data/music/fmlink";
)
func main()  {
	if(len(os.Args) <= 1){
		fmt.Println("请输入网易音乐链接.");
		return ;
	}
	fmt.Println("fetching msg from ",os.Args[1]);

	nurl := strings.Replace(os.Args[1],"#/","",-1);

	uri,err := url.Parse(nurl);

	if(err != nil){
		fmt.Println("解析URL时出错：",err);
		return ;
	}
	response,err := http.Get(uri.String());
	if(err != nil){
		fmt.Println("获取远程URL内容时出错：",err);
		return ;
	}

	responseBody,err := ioutil.ReadAll(response.Body);
	response.Body.Close();
	if(err != nil){
		fmt.Println("读取远程URL响应内容时出错：",err);
		return ;
	}

	var path string;

	if  os.IsPathSeparator('\\') {
		path = "\\";
	}else{
		path = "/";
	}
	dir, _ := os.Getwd();

	dir = dir +path+"songs_dir";

	if _,err := os.Stat(dir);err != nil{
		err = os.Mkdir(dir, os.ModePerm);
		if(err != nil){
			fmt.Println("创建目录失败：",err);
			return ;
		}
	}

	reg := regexp.MustCompile(`<ul class="f-hide">(.*?)</ul>`);

	mm := reg.FindAllString(string(responseBody),-1);

	waitGroup := sync.WaitGroup{};

	if(len(mm) > 0){
		reg = regexp.MustCompile(`<li><a .*?>(.*?)</a></li>`);


		contents := mm[0];
		urlli := reg.FindAllSubmatch([]byte(contents),-1);

		for _,item := range urlli{

			murl,_ := url.Parse(SuggestionUrl);

			query := murl.Query();
			query.Set("word", string(item[1]));
			query.Set("version","2");
			query.Set("from","0");

			murl.RawQuery = query.Encode();

			res,err := http.Get(murl.String());
			if(err != nil){
				fmt.Println("获取",murl,"出错：",err);
				continue;
			}
			content,err := ioutil.ReadAll(res.Body);
			res.Body.Close();
			if(err != nil){
				fmt.Println("解析",murl,"响应值时出错：",err);
				continue;
			}

			var dat map[string]interface{};

			err = json.Unmarshal(content, &dat);

			if err != nil {
				fmt.Println("反序列化JSON时出错:",err);
				continue;
			}

			if _,ok := dat["data"]; ok == false{
				fmt.Println("没有找到音乐资源:",string(item[1]));
				continue;
			}

			songid := dat["data"].(map[string]interface{})["song"].([]interface{})[0].(map[string]interface{})["songid"].(string);

			link ,err:= url.Parse(Fmlink);
			if(err != nil){
				fmt.Println("解析音乐链接时出错：",err);
				continue;
			}
			query = link.Query();
			query.Set("songIds",songid);
			query.Set("type","flac");
			link.RawQuery = query.Encode();
			res ,err = http.Get(link.String());

			if(err != nil){
				fmt.Println("获取音乐文件时出错：",err);
				continue;
			}

			content,err = ioutil.ReadAll(res.Body);
			res.Body.Close();
			if(err != nil){
				fmt.Println("读取音乐文件数据时出错：",err);
				continue;
			}

			var data map[string]interface{};

			err = json.Unmarshal(content,&data);

			if code,ok:= data["errorCode"]; ((ok && code.(float64) == 22005) || err != nil){
				fmt.Println("解析音乐文件时出错：",err,string(content));
				continue;
			}

			songlink := data["data"].(map[string]interface{})["songList"].([]interface{})[0].(map[string]interface{})["songLink"].(string);

			r := []rune(songlink)
			if(len(r) < 10){
				fmt.Println("没有无损音乐地址");
				continue;
			}

			songname :=  data["data"].(map[string]interface{})["songList"].([]interface{})[0].(map[string]interface{})["songName"].(string);

			artistName :=  data["data"].(map[string]interface{})["songList"].([]interface{})[0].(map[string]interface{})["artistName"].(string);

			filename := dir + path + songname+"-"+artistName+".flac"

			waitGroup.Add(1);
			go func() {
				fmt.Println(songname , " is downloading now ......");

				songRes ,err:= http.Get(songlink);
				if(err != nil){
					fmt.Println("下载文件时出错：",songlink);
					waitGroup.Done();
					return ;
				}

				songFile,err := os.Create(filename);
				written,err := io.Copy(songFile,songRes.Body);
				if(err != nil){
					fmt.Println("保存音乐文件时出错：",err);
					waitGroup.Done();
					return ;
				}
				fmt.Println("下载",filename,"完成,文件大小：",fmt.Sprintf("%.2f", (float64(written)/(1024*1024))),"MB");
				waitGroup.Done();
			}();

		}

	}
	waitGroup.Wait();
}
