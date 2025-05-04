for i in {1..100}; do
    echo "Run #$i"
    #go test -v -test.run TestGraphAddToObjec
    #go test -v -test.run TestGraphStringArrayLitteral
	#go test -v -test.run TestGraphMergeLitteral
    go test -v 
    if [ $? -ne 0 ]; then
        echo "Test failed on run #$i"
        break
    fi
done
