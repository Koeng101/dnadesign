import os
import platform
from packaging import tags
from setuptools import setup, find_packages

def get_shared_lib_ext():
    if platform.system() == "Darwin":
        return ".dylib"
    elif platform.system() == "Windows":
        return ".dll"
    else:
        return ".so"


def get_platform_tag():
    if sys.platform.startswith('linux'):
        return f'manylinux_2_17_{platform.machine()}'
    elif sys.platform.startswith('darwin'):
        if platform.machine() == 'arm64':
            return f'macosx_11_0_{platform.machine()}'
        else:
            return f'macosx_10_9_{platform.machine()}'
    elif sys.platform.startswith('win'):
        return 'win_amd64'
    else:
        return 'any'

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

    plat_name=f"{platform_tag.platform}_{platform_tag.abi}_{platform_tag.interpreter}_{platform_tag.implementation}",
    include_package_data=True,
    zip_safe=False,
    author='Keoni Gandall',
    author_email='koeng101@gmail.com',
    description='Python bindings for dnadesign',
    url='https://github.com/koeng101/dnadesign'
)
