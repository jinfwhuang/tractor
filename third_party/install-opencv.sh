#!/bin/bash

. ../setup.bash

cd $FARM_NG_ROOT/third_party
mkdir -p build-opencv
cd build-opencv

cmake \
    -DCMAKE_INSTALL_PREFIX=$FARM_NG_ROOT/env \
    -DCMAKE_PREFIX_PATH=$FARM_NG_ROOT/env/ \
    -DCMAKE_BUILD_TYPE=Release \
    -D INSTALL_PYTHON_EXAMPLES=OFF \
    -D INSTALL_C_EXAMPLES=OFF \
    -D PYTHON_EXECUTABLE=$(which python) \
    -D BUILD_opencv_python2=OFF \
    -D PYTHON3_EXECUTABLE=$(which python) \
    -D WITH_GSTREAMER=ON \
    -D BUILD_EXAMPLES=OFF \
    -D BUILD_TESTS=OFF \
    ../opencv

make -j$(nproc --ignore=1)
make install
