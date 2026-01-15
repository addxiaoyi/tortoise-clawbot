from setuptools import setup, find_packages

setup(
    name="tortoise-sdk",
    version="0.1.0",
    description="Tortoise AI Agent Framework Python SDK",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    author="Tortoise Team",
    author_email="team@tortoise.ai",
    url="https://github.com/tortoise-ai/tortoise",
    packages=find_packages(exclude=["tests", "tests.*"]),
    install_requires=[
        "aiohttp>=3.9.0",
        "websockets>=12.0",
    ],
    extras_require={
        "dev": [
            "pytest>=7.4.0",
            "pytest-asyncio>=0.23.0",
            "pytest-cov>=4.1.0",
            "black>=23.0.0",
            "mypy>=1.7.0",
        ],
    },
    python_requires=">=3.8",
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: Apache Software License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
    ],
    keywords="ai agent chatbot llm openai anthropic",
    license="Apache-2.0",
)
