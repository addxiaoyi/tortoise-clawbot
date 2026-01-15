cmake_minimum_required(VERSION 3.15)
project(tortoise C)

set(TORTOISE_VERSION_MAJOR 0)
set(TORTOISE_VERSION_MINOR 1)
set(TORTOISE_VERSION_PATCH 0)

option(BUILD_SHARED_LIBS "Build shared libraries" ON)
option(BUILD_EXAMPLES "Build examples" ON)

set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)

# Warning flags
if(MSVC)
    add_compile_options(/W4)
else()
    add_compile_options(-Wall -Wextra -pedantic)
endif()

# Source files
set(TORTOISE_SOURCES
    src/tortoise.c
)

# Header files
set(TORTOISE_HEADERS
    include/tortoise.h
)

# Create library
add_library(tortoise ${TORTOISE_SOURCES} ${TORTOISE_HEADERS})

# Include directories
target_include_directories(tortoise PUBLIC
    ${CMAKE_CURRENT_SOURCE_DIR}/include
)

# Set version
set_target_properties(tortoise PROPERTIES
    VERSION ${TORTOISE_VERSION_MAJOR}.${TORTOISE_VERSION_MINOR}.${TORTOISE_VERSION_PATCH}
    SOVERSION ${TORTOISE_VERSION_MAJOR}
)

# Export symbols for Windows
if(WIN32)
    set_target_properties(tortoise PROPERTIES
        DEFINE_SYMBOL TORTOISE_EXPORTS
    )
endif()

# Build examples
if(BUILD_EXAMPLES)
    add_executable(example examples/example.c)
    target_link_libraries(example tortoise)
    target_include_directories(example PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/include
    )
endif()

# Testing
enable_testing()
add_test(NAME tortoise_test COMMAND example)

# Installation
install(TARGETS tortoise
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
    RUNTIME DESTINATION bin
)

install(DIRECTORY include/
    DESTINATION include
)

# Pkg-config
configure_file(tortoise.pc.in tortoise.pc @ONLY)
install(FILES ${CMAKE_BINARY_DIR}/tortoise.pc
    DESTINATION pkg-config
)
