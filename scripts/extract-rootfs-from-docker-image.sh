docker build -t samir-os:1 .
cid=$(docker create samir-os:1)
docker export "$cid" -o rootfs.tar
docker rm "$cid"
mkdir rootfs
tar -xf rootfs.tar -C rootfs
