#/bin/bash 
docker build -t colonies-builder .. -f Dockerfile.build_ubuntu_2020
containerid=$(docker run -d colonies-builder /bin/bash)
docker cp $containerid:/build/lib/libcryptolib.so ../lib/ubuntu_2020
