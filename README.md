# GOALO

### Docker

#### Build Image
> docker build . -t goalo:\<tag\>

#### Run Image
> docker run -it -v \<host_path\>/data/:/app/data -v \<host_path\>/config:/app/config/ goalo:\<tag\>
