for i in {1..1000}; do
    echo "Run #$i"
    go test -v 
    if [ $? -ne 0 ]; then
        echo "Test failed on run #$i"
        break
    fi
done
