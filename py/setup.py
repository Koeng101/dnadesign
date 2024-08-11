import os
import platform
from setuptools import setup, find_packages

def get_shared_lib_ext():
    if platform.system() == "Darwin":
        return ".dylib"
    elif platform.system() == "Windows":
        return ".dll"
    else:
        return ".so"

setup(
    name='dnadesign',
    version='0.1.1',
    packages=find_packages(),
    package_data={'dnadesign': ['definitions.h', 'libdnadesign.h', "libdnadesign" + get_shared_lib_ext()]},
    install_requires=[
        "cffi>=1.0.0",
    ],
    setup_requires=[
        "cffi>=1.0.0",
    ],

    include_package_data=True,
    zip_safe=False,
    author='Keoni Gandall',
    author_email='koeng101@gmail.com',
    description='Python bindings for dnadesign',
    url='https://github.com/koeng101/dnadesign'
)
