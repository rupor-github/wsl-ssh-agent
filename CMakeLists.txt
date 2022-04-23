cmake_minimum_required(VERSION 3.0)

message(STATUS "Running cmake version ${CMAKE_VERSION}")
if (NOT CMAKE_TOOLCHAIN_FILE)
    set(CMAKE_TOOLCHAIN_FILE "../cmake/win64.toolchain" CACHE STRING "default toolchain" FORCE)
endif()
if (NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE "Release" CACHE STRING "default to release build type" FORCE)
endif()

project(wsl-ssh-agent NONE)
set(CMAKE_MODULE_PATH ${CMAKE_MODULE_PATH} ${CMAKE_SOURCE_DIR}/cmake)

message(STATUS "Build type: ${CMAKE_BUILD_TYPE}")
message(STATUS "Generator: ${CMAKE_GENERATOR}")

# Release installation directory
if(NOT CMAKE_CROSSCOMPILING)
    message(FATAL_ERROR "Currently unsuppored environment. CMake willl exit now.")
else()
    set(HOST_GOOS linux)
    set(HOST_GOARCH amd64)
endif()

if(MSYSTEM_NAME STREQUAL "win32")
    set(CROSS_GOOS windows)
    set(CROSS_GOARCH 386)
    message(STATUS "Targeting 32 bits Windows executable")
elseif(MSYSTEM_NAME STREQUAL "win64")
    set(CROSS_GOOS windows)
    set(CROSS_GOARCH amd64)
    message(STATUS "Targeting 64 bits Windows executable")
else()
    message(FATAL_ERROR "Currently unsuppored environment \"${MSYSTEM_NAME}\". CMake willl exit now.")
endif()

set(DEST_DIR "${PROJECT_SOURCE_DIR}/bin")
message(STATUS "Installation directory: ${DEST_DIR}")

set(GO_MIN_REQURED_VERSION 1.17)
find_package(Go ${GO_MIN_REQURED_VERSION} REQUIRED)
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

# Project version number
set(PRJ_VERSION_Major "1")
set(PRJ_VERSION_Minor "6")
set(PRJ_VERSION_Patch "0")

if (EXISTS "${PROJECT_SOURCE_DIR}/.git" AND IS_DIRECTORY "${PROJECT_SOURCE_DIR}/.git")
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
configure_file("${PROJECT_SOURCE_DIR}/cmake/agent.rc.in" "${PROJECT_SOURCE_DIR}/cmd/agent/resources.rc")
configure_file("${PROJECT_SOURCE_DIR}/cmake/agent.xml.in" "${PROJECT_SOURCE_DIR}/cmd/agent/manifest.xml")
configure_file("${PROJECT_SOURCE_DIR}/cmake/wsl-ssh-agent.json.in" "${PROJECT_SOURCE_DIR}/wsl-ssh-agent.json")

# for user convinience
execute_process(COMMAND ${CMAKE_COMMAND} -E copy ${CMAKE_CURRENT_SOURCE_DIR}/../../jstarks/npiperelay/npiperelay.exe ${PROJECT_BINARY_DIR})

# distribute history
execute_process(COMMAND ${CMAKE_COMMAND} -E env ${GIT_EXECUTABLE} changelog --all --tag "v${PRJ_VERSION_Major}.${PRJ_VERSION_Minor}.${PRJ_VERSION_Patch}" --stdout OUTPUT_FILE ${PROJECT_BINARY_DIR}/changelog.txt ERROR_QUIET)

message(STATUS "Building version \"${PRJ_VERSION_Major}.${PRJ_VERSION_Minor}.${PRJ_VERSION_Patch}\" git \"${GIT_HASH}\"")

if(CMAKE_BUILD_TYPE STREQUAL "Debug")
    # Create git pre-commit hook to check formatting before commit
    if(EXISTS "${PROJECT_SOURCE_DIR}/.git" AND IS_DIRECTORY "${PROJECT_SOURCE_DIR}/.git")
        execute_process(COMMAND ${CMAKE_COMMAND} -E copy ${PROJECT_SOURCE_DIR}/cmake/pre-commit ${PROJECT_SOURCE_DIR}/.git/hooks/. WORKING_DIRECTORY ${PROJECT_SOURCE_DIR})
    endif()
endif()

########################################################################################################
# Some contexts dependent settings
########################################################################################################

list(APPEND GO_ENV
    GOPATH=${GO_PATH}
    GOOS=${CROSS_GOOS}
    GOARCH=${CROSS_GOARCH}
    )

if(CMAKE_BUILD_TYPE STREQUAL "Debug")
    set(GO_ARGS "-mod=mod")
else()
    set(GO_ARGS "-mod=vendor")
endif()

########################################################################################################
# main target - always out of date, "go build" will figure it out
########################################################################################################

add_custom_target(release
    DEPENDS ${PROJECT_SOURCE_DIR}/wsl-ssh-agent.zip
        ${PROJECT_SOURCE_DIR}/wsl-ssh-agent.json
    COMMAND ${CMAKE_SOURCE_DIR}/cmake/sign.sh
    COMMENT "Building release..."
    WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}")

add_custom_command(OUTPUT ${PROJECT_SOURCE_DIR}/wsl-ssh-agent.zip
    DEPENDS ${PROJECT_BINARY_DIR}/changelog.txt
        ${PROJECT_BINARY_DIR}/npiperelay${CMAKE_EXECUTABLE_SUFFIX}
        ${PROJECT_BINARY_DIR}/wsl-ssh-agent-gui${CMAKE_EXECUTABLE_SUFFIX}
    COMMAND ${CMAKE_COMMAND} -E tar "cfv" ${PROJECT_SOURCE_DIR}/wsl-ssh-agent.zip --format=zip
        changelog.txt npiperelay${CMAKE_EXECUTABLE_SUFFIX} wsl-ssh-agent-gui${CMAKE_EXECUTABLE_SUFFIX}
    COMMENT "Archiving release..."
    WORKING_DIRECTORY "${PROJECT_BINARY_DIR}")

add_custom_target(bin_agent ALL
    DEPENDS ${PROJECT_SOURCE_DIR}/cmd/agent/resources.syso
    BYPRODUCTS ${PROJECT_BINARY_DIR}/wsl-ssh-agent-gui${CMAKE_EXECUTABLE_SUFFIX}
    COMMAND ${GO_ENV} ${GO_EXECUTABLE} build -trimpath -o ${PROJECT_BINARY_DIR}/wsl-ssh-agent-gui${CMAKE_EXECUTABLE_SUFFIX}
        -ldflags='-H=windowsgui'
        ${GO_ARGS}
        ./cmd/agent
    COMMENT "Building wsl-ssh-agent-gui..."
    WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}")

add_custom_command(OUTPUT ${PROJECT_SOURCE_DIR}/cmd/agent/resources.syso
     DEPENDS ${PROJECT_SOURCE_DIR}/cmd/agent/resources.rc
         ${PROJECT_SOURCE_DIR}/cmd/agent/manifest.xml
         ${PROJECT_SOURCE_DIR}/cmd/agent/icon.ico
     COMMAND ${CMAKE_RC_COMPILER} -O coff
         -o ${PROJECT_SOURCE_DIR}/cmd/agent/resources.syso
         -i ${PROJECT_SOURCE_DIR}/cmd/agent/resources.rc
     WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}"
     COMMENT "Building wsl-ssh-agent-gui resources...")

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
        COMMAND ${GO_ENV} ${GO_EXECUTABLE} mod tidy -compat=${GO_MIN_REQURED_VERSION}
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
