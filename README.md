# 网易云音乐无损音乐下载

根据网易云音乐歌单, 下载对应无损FLAC歌曲到本地.

该程序是根据Python版本使用Golang重写，原版位于：https://github.com/YongHaoWu/NeteaseCloudMusicFlac

使用方法：

    go build //编译
    go install //安装
    NeteaseCloudMusicFlac.exe http://music.163.com/#/playlist?id=145258012 //解析并下载。音乐会下载到当前程序目录的songs_dir目录下。

#### 本程序仅供学习之用。
