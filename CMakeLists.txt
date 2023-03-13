cmake_minimum_required(VERSION 3.18)

project(wsl-ssh-agent NONE)

message(STATUS "Running cmake version ${CMAKE_VERSION}")
if (NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE "Release" CACHE STRING "default to release build type" FORCE)
endif()
set_property(CACHE CMAKE_BUILD_TYPE PROPERTY STRINGS "Debug" "Release")
message(STATUS "Build type: ${CMAKE_BUILD_TYPE}")
message(STATUS "Generator: ${CMAKE_GENERATOR}")

set(CMAKE_MODULE_PATH ${CMAKE_MODULE_PATH} ${CMAKE_SOURCE_DIR}/cmake)

if (NOT "$ENV{MSYSTEM_NAME}" STREQUAL "")
    set(MSYSTEM_NAME "$ENV{MSYSTEM_NAME}" CACHE INTERNAL "Copied from environment")
else()
    set(MSYSTEM_NAME "linux_amd64" CACHE INTERNAL "Host environment")
endif()

if(MSYSTEM_NAME STREQUAL "win64")
    set(CROSS_GOOS windows)
    set(CROSS_GOARCH amd64)
     # we are not using cmake cross compilation abilities
    set(CROSS_EXECUTABLE_SUFFIX ".exe")
    find_program(CROSS_RC_COMPILER x86_64-w64-mingw32-windres PATHS usr PATH_SUFFIXES bin DOC "Windows resource compiler path" REQUIRED)
    message(STATUS "Targeting 64 bits Windows executable")
else()
    message(FATAL_ERROR "Currently unsuppored environment \"${MSYSTEM_NAME}\".")
endif()

find_package(Go REQUIRED)
find_package(Git REQUIRED)

# Make sure we setup go properly
if (DEFINED ENV{GOPATH})
    set(GO_PATH "$ENV{GOPATH}")
else()
    set(GO_PATH "/tmp/gopkg_${CMAKE_PROJECT_NAME}")
    execute_process(COMMAND ${CMAKE_COMMAND} -E make_directory ${GO_PATH})
endif()
message(STATUS "Using go modules with GOPATH=${GO_PATH}")
if(GO_PATH MATCHES "^${PROJECT_SOURCE_DIR}.*")
    message(FATAL_ERROR "We are working in \"go modules outside of GOPATH mode\" - GOPATH should be outside of project source directory!")
endif()

########################################################################################################
# Setting up versioning
########################################################################################################

# Project version number - comes from release version tag
set(GIT_REF $ENV{GITHUB_REF})
if(GIT_REF MATCHES "refs/tags/v([0-9]+)\\.([0-9]+)\\.([0-9]+)")
    set(PRJ_VERSION_Major ${CMAKE_MATCH_1})
    set(PRJ_VERSION_Minor ${CMAKE_MATCH_2})
    set(PRJ_VERSION_Patch ${CMAKE_MATCH_3})
else()
    set(PRJ_VERSION_Major 0)
    set(PRJ_VERSION_Minor 0)
    set(PRJ_VERSION_Patch 0)
endif()

if(DEFINED ENV{GITHUB_SHA})
    set(GIT_HASH "$ENV{GITHUB_SHA}")
elseif(EXISTS "${PROJECT_SOURCE_DIR}/.git" AND IS_DIRECTORY "${PROJECT_SOURCE_DIR}/.git")
    execute_process(COMMAND ${CMAKE_SOURCE_DIR}/cmake/githash.sh ${GIT_EXECUTABLE}
        WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}"
        RESULT_VARIABLE res
        OUTPUT_VARIABLE GIT_HASH
        ERROR_QUIET
        OUTPUT_STRIP_TRAILING_WHITESPACE)
    # This will cause reconfiguration on every commit
    set_property(DIRECTORY APPEND PROPERTY CMAKE_CONFIGURE_DEPENDS "${PROJECT_SOURCE_DIR}/.git/index")
else()
    set(GIT_HASH "no-git")
endif()
configure_file("${PROJECT_SOURCE_DIR}/cmake/version.go.in" "${PROJECT_SOURCE_DIR}/misc/version.go")

if(CMAKE_BUILD_TYPE STREQUAL "Debug")
    # Create local git hooks
    if(EXISTS "${PROJECT_SOURCE_DIR}/.git" AND IS_DIRECTORY "${PROJECT_SOURCE_DIR}/.git")
        execute_process(COMMAND ${CMAKE_COMMAND} -E copy ${PROJECT_SOURCE_DIR}/cmake/pre-commit ${PROJECT_SOURCE_DIR}/.git/hooks/. WORKING_DIRECTORY ${PROJECT_SOURCE_DIR})
        configure_file("${PROJECT_SOURCE_DIR}/cmake/pre-push" "${PROJECT_SOURCE_DIR}/.git/hooks/pre-push" @ONLY)
    endif()
endif()

if(CROSS_GOOS STREQUAL "windows")
    if(NOT EXISTS "${PROJECT_SOURCE_DIR}/gclpr.json")
        configure_file("${PROJECT_SOURCE_DIR}/cmake/wsl-ssh-agent.json.in" "${PROJECT_SOURCE_DIR}/wsl-ssh-agent.json" @ONLY)
    endif()
    configure_file("${PROJECT_SOURCE_DIR}/cmake/agent.rc.in" "${PROJECT_SOURCE_DIR}/cmd/gui/resources.rc" @ONLY)
    configure_file("${PROJECT_SOURCE_DIR}/cmake/agent.xml.in" "${PROJECT_SOURCE_DIR}/cmd/gui/manifest.xml" @ONLY)
endif()

message(STATUS "Building version \"${PRJ_VERSION_Major}.${PRJ_VERSION_Minor}.${PRJ_VERSION_Patch}\" git \"${GIT_HASH}\"")

########################################################################################################
# Some contexts dependent settings
########################################################################################################

list(APPEND GO_ENV
    GOPATH=${GO_PATH}
    GOOS=${CROSS_GOOS}
    GOARCH=${CROSS_GOARCH}
    CGO_ENABLED=0
    )

if(CMAKE_BUILD_TYPE STREQUAL "Debug")
    set(TEST_RUN_ARGS "-v")
    set(GO_MOD_ARGS "-mod=mod")
else()
    set(GO_MOD_ARGS "-mod=vendor")
endif()

########################################################################################################
# main target - always out of date, "go build" will figure it out
########################################################################################################

add_custom_target(release
    DEPENDS ${PROJECT_SOURCE_DIR}/wsl-ssh-agent.zip
        ${PROJECT_SOURCE_DIR}/wsl-ssh-agent.json
    COMMAND ${CMAKE_SOURCE_DIR}/cmake/scoophash.sh ${MSYSTEM_NAME}
    COMMENT "Building release..."
    WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}")

add_custom_command(OUTPUT ${PROJECT_SOURCE_DIR}/wsl-ssh-agent.zip
    DEPENDS ${PROJECT_BINARY_DIR}/npiperelay${CROSS_EXECUTABLE_SUFFIX}
        ${PROJECT_BINARY_DIR}/wsl-ssh-agent-gui${CROSS_EXECUTABLE_SUFFIX}
    COMMAND ${CMAKE_COMMAND} -E tar "cfv" ${PROJECT_SOURCE_DIR}/wsl-ssh-agent.zip --format=zip
        npiperelay${CROSS_EXECUTABLE_SUFFIX} wsl-ssh-agent-gui${CROSS_EXECUTABLE_SUFFIX}
    COMMENT "Archiving release..."
    WORKING_DIRECTORY "${PROJECT_BINARY_DIR}")

add_custom_target(bin_agent ALL
    DEPENDS ${PROJECT_SOURCE_DIR}/cmd/agent/resources.syso
    BYPRODUCTS ${PROJECT_BINARY_DIR}/wsl-ssh-agent-gui${CROSS_EXECUTABLE_SUFFIX}
    COMMAND ${GO_ENV} ${GO_EXECUTABLE} build ${GO_MOD_ARGS} -trimpath -o ${PROJECT_BINARY_DIR}/wsl-ssh-agent-gui${CROSS_EXECUTABLE_SUFFIX}
        -ldflags='-H=windowsgui'
        ./cmd/agent
    COMMENT "Building wsl-ssh-agent-gui..."
    WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}")

add_custom_command(OUTPUT ${PROJECT_SOURCE_DIR}/cmd/agent/resources.syso
    DEPENDS ${PROJECT_SOURCE_DIR}/cmd/agent/resources.rc
        ${PROJECT_SOURCE_DIR}/cmd/agent/manifest.xml
        ${PROJECT_SOURCE_DIR}/cmd/agent/icon.ico
    COMMAND ${CROSS_RC_COMPILER} -O coff
        -o ${PROJECT_SOURCE_DIR}/cmd/agent/resources.syso
        -i ${PROJECT_SOURCE_DIR}/cmd/agent/resources.rc
    WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}"
    COMMENT "Building wsl-ssh-agent-gui resources...")

add_custom_target(bin_npiperelay
    BYPRODUCTS ${PROJECT_BINARY_DIR}/npiperelay${CROSS_EXECUTABLE_SUFFIX}
    COMMAND ${GO_ENV} ${GO_EXECUTABLE} build ${GO_MOD_ARGS} -trimpath -o ${PROJECT_BINARY_DIR}/npiperelay${CROSS_EXECUTABLE_SUFFIX}
        github.com/jstarks/npiperelay
    COMMENT "Building npiperelay..."
    WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}")

########################################################################################################
# Development
########################################################################################################

if(CMAKE_BUILD_TYPE STREQUAL "Debug")

    ########################################################################################################
    # Go linter
    ########################################################################################################

    add_custom_command(OUTPUT ${PROJECT_BINARY_DIR}/staticcheck
        COMMAND GOPATH=${GO_PATH} GOOS= GOARCH= ${GO_EXECUTABLE} build ${GO_MOD_ARGS} -o ${PROJECT_BINARY_DIR}/staticcheck
            honnef.co/go/tools/cmd/staticcheck
        WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}"
        COMMENT "Building staticcheck executable...")

    add_custom_target(lint
        DEPENDS ${PROJECT_BINARY_DIR}/staticcheck
        COMMAND ${PROJECT_BINARY_DIR}/staticcheck -f stylish -tests=false ./...
        COMMENT "Linting project..."
        WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}")

    ########################################################################################################
    # Go dependency management
    ########################################################################################################

    add_custom_target(deps_tidy
        SOURCES ${PROJECT_SOURCE_DIR}/go.mod
        COMMAND ${GO_ENV} ${GO_EXECUTABLE} mod tidy -go=${GO_LANG_VERSION}
        WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}"
        COMMENT "Cleaning dependencies (go mod tidy)"
        VERBATIM)

    add_custom_target(deps_download
        SOURCES ${PROJECT_SOURCE_DIR}/go.mod
        COMMAND ${GO_ENV} ${GO_EXECUTABLE} mod download all
        WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}"
        COMMENT "Downloading dependencies (go mod download)"
        VERBATIM)

    add_custom_target(deps_vendor
        SOURCES ${PROJECT_SOURCE_DIR}/go.mod
        COMMAND ${GO_ENV} ${GO_EXECUTABLE} mod vendor
        WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}"
        COMMENT "Creating vendor directory (go mod vendor)"
        VERBATIM)

endif()
