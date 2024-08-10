from setuptools import setup, find_packages
from setuptools.command.install import install
import platform
import os

class CustomInstall(install):
    def run(self):
        # Install the appropriate shared library based on the user's platform
        system = platform.system().lower()
        arch = 'amd64' if platform.machine() in ['x86_64', 'AMD64'] else 'arm64'
        lib_file = f'dnadesign/lib/libdnadesign_{system}_{arch}.so'
        if not os.path.exists(lib_file):
            raise FileNotFoundError(f"Could not find required library: {lib_file}")
        os.rename(lib_file, 'dnadesign/libdnadesign.so')
        install.run(self)

setup(
    name='dnadesign',
    version='0.1.0',
    packages=find_packages(),
    cmdclass={'install': CustomInstall},
    package_data={'dnadesign': ['libdnadesign_*.so']},
    include_package_data=True,
    zip_safe=False
)

