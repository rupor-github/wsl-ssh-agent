cmake_minimum_required(VERSION 3.0)

project(wsl-ssh-agent NONE)

message(STATUS "Running cmake version ${CMAKE_VERSION}")
if (NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE "Release" CACHE STRING "default to release build type" FORCE)
endif()
set_property(CACHE CMAKE_BUILD_TYPE PROPERTY STRINGS "Debug" "Release")
message(STATUS "Build type: ${CMAKE_BUILD_TYPE}")
message(STATUS "Generator: ${CMAKE_GENERATOR}")

if(CMAKE_BUILD_TYPE STREQUAL "Debug")
    message(STATUS "********************************************************************************")
    message(STATUS "*** Do not forget to do \"make deps_vendor\" before switching to Release mode! ***")
    message(STATUS "********************************************************************************")
endif()

set(CMAKE_MODULE_PATH ${CMAKE_MODULE_PATH} ${CMAKE_SOURCE_DIR}/cmake)

# Release installation directory
if(NOT CMAKE_CROSSCOMPILING)
    message(FATAL_ERROR "Currently unsuppored environment. CMake willl exit now.")
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

find_package(Go 1.13 REQUIRED)
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
set(PRJ_VERSION_MAJOR 1)
set(PRJ_VERSION_MINOR 1)

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

# distribute history
execute_process(COMMAND ${CMAKE_COMMAND} -E env ${GIT_EXECUTABLE} changelog --all --tag "${PRJ_VERSION_MAJOR}.${PRJ_VERSION_MINOR}" --stdout OUTPUT_FILE ${PROJECT_BINARY_DIR}/changelog.txt ERROR_QUIET)
install(FILES ${PROJECT_BINARY_DIR}/changelog.txt DESTINATION ${DEST_DIR} CONFIGURATIONS Release)

message(STATUS "Building version \"${PRJ_VERSION_MAJOR}.${PRJ_VERSION_MINOR}\" git \"${GIT_HASH}\"")

if(CMAKE_BUILD_TYPE STREQUAL "Debug")
    # Create git pre-commit hook to check formatting before commit
    if(EXISTS "${PROJECT_SOURCE_DIR}/.git" AND IS_DIRECTORY "${PROJECT_SOURCE_DIR}/.git")
        execute_process(COMMAND ${CMAKE_COMMAND} -E copy ${PROJECT_SOURCE_DIR}/cmake/pre-commit ${PROJECT_SOURCE_DIR}/.git/hooks/. WORKING_DIRECTORY ${PROJECT_SOURCE_DIR})
    endif()
endif()
if(EXISTS "${PROJECT_SOURCE_DIR}/${KINDLEGEN}")
    execute_process(COMMAND ${CMAKE_COMMAND} -E copy ${PROJECT_SOURCE_DIR}/${KINDLEGEN} ${PROJECT_BINARY_DIR}/.)
endif()

########################################################################################################
# Some contexts dependent settings
########################################################################################################

if(CMAKE_BUILD_TYPE STREQUAL "Debug")
    set(TEST_RUN_ARGS "-v")
else()
    set(GO_MOD_ARGS "-mod=vendor")
endif()

########################################################################################################
# main target - always out of date, "go build" will figure it out
########################################################################################################

add_custom_target(bin_agent ALL
    DEPENDS ${PROJECT_SOURCE_DIR}/static/files.go
    COMMAND GOPATH=${GO_PATH} GO111MODULE=on GOOS=${CROSS_GOOS} GOARCH=${CROSS_GOARCH}
        ${GO_EXECUTABLE} build -trimpath -o ${PROJECT_BINARY_DIR}/wsl-ssh-agent-gui${CMAKE_EXECUTABLE_SUFFIX}
        -ldflags='-X main.LastGitCommit=${GIT_HASH} -H=windowsgui'
        ${GO_MOD_ARGS}
    COMMENT "Building wsl-ssh-agent-gui..."
    WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}")
install(PROGRAMS ${PROJECT_BINARY_DIR}/wsl-ssh-agent-gui${CMAKE_EXECUTABLE_SUFFIX} DESTINATION ${DEST_DIR} CONFIGURATIONS Release)

########################################################################################################
# Development
########################################################################################################

if(CMAKE_BUILD_TYPE STREQUAL "Debug")

    ########################################################################################################
    # Code generation - on host platform only
    ########################################################################################################

    add_custom_command(OUTPUT ${PROJECT_BINARY_DIR}/go-bindata
        COMMAND GOPATH=${GO_PATH} GO111MODULE=on
            ${GO_EXECUTABLE} build -o ${PROJECT_BINARY_DIR}/go-bindata
            github.com/shuLhan/go-bindata/cmd/go-bindata
        WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}"
        COMMENT "Building go-bindata executable...")
    add_custom_target(bin_go-bindata
        DEPENDS ${PROJECT_BINARY_DIR}/go-bindata)

    file(GLOB_RECURSE SRCS_STATIC LIST_DIRECTORIES false ${PROJECT_SOURCE_DIR}/static/*)
    list(REMOVE_ITEM SRCS_STATIC ${PROJECT_SOURCE_DIR}/static/files.go ${PROJECT_SOURCE_DIR}/static/doc.go)
    add_custom_command(OUTPUT ${PROJECT_SOURCE_DIR}/static/files.go
        DEPENDS ${PROJECT_BINARY_DIR}/go-bindata ${SRCS_STATIC}
        # -nocompress
        COMMAND GOPATH=${GO_PATH}
            ${PROJECT_BINARY_DIR}/go-bindata
                -prefix ${PROJECT_SOURCE_DIR}/static/
                -ignore static/files.go
                -ignore static/doc.go
                -pkg static
                -o static/files.go
                ${PROJECT_SOURCE_DIR}/static/...
        WORKING_DIRECTORY ${PROJECT_SOURCE_DIR}
        COMMENT "Generating Go resources for static assets")

    ########################################################################################################
    # Go linter
    ########################################################################################################

    # Running linters on main target
    find_program(GO_LINTER golangci-lint)
    if(NOT GO_LINTER)
        set(golangci-lint-ver "1.17.1")
        message(STATUS "Preparing golangci-lint ${golangci-lint-ver}")
        file(DOWNLOAD "https://github.com/golangci/golangci-lint/releases/download/v${golangci-lint-ver}/golangci-lint-${golangci-lint-ver}-linux-amd64.tar.gz" ${PROJECT_BINARY_DIR}/golangci-lint.tar.gz
            INACTIVITY_TIMEOUT 60
            TIMEOUT 300
            STATUS DOWNLOAD_RES
            EXPECTED_HASH SHA256=62693d3351858206af0bfe975f7f79a8ac9124502e8de7906f377c231f37f7d3)
        list(GET DOWNLOAD_RES 0 RES)
        list(GET DOWNLOAD_RES 1 MSG)
        if(RES)
            message(FATAL_ERROR, ${MSG})
        endif()
        execute_process(COMMAND ${CMAKE_COMMAND} -E tar xf golangci-lint.tar.gz WORKING_DIRECTORY ${PROJECT_BINARY_DIR})
        execute_process(COMMAND ${CMAKE_COMMAND} -E copy golangci-lint-${golangci-lint-ver}-linux-amd64/golangci-lint golangci-lint WORKING_DIRECTORY ${PROJECT_BINARY_DIR})
        execute_process(COMMAND ${CMAKE_COMMAND} -E remove_directory golangci-lint-${golangci-lint-ver}-linux-amd64 WORKING_DIRECTORY ${PROJECT_BINARY_DIR})
        execute_process(COMMAND ${CMAKE_COMMAND} -E remove golangci-lint.tar.gz WORKING_DIRECTORY ${PROJECT_BINARY_DIR})
        set(GO_LINTER ${PROJECT_BINARY_DIR}/golangci-lint)
    endif()

    add_custom_target(lint
        COMMAND GOPATH=${GO_PATH} GO111MODULE=on GOOS=${CROSS_GOOS} GOARCH=${CROSS_GOARCH} ${GO_LINTER} run
        # COMMAND GOPATH=${GO_PATH} GO111MODULE=on ${GO_LINTER} run
        COMMENT "Mega-linting project with ${GO_LINTER}..."
        WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}")

    ########################################################################################################
    # Go dependency management
    ########################################################################################################

    add_custom_target(deps_tidy
        SOURCES ${PROJECT_SOURCE_DIR}/go.mod
        COMMAND GOPATH=${GO_PATH} GO111MODULE=on GOOS=${CROSS_GOOS} GOARCH=${CROSS_GOARCH} ${GO_EXECUTABLE} mod tidy
        WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}"
        COMMENT "Cleaning dependencies (go mod tidy)"
        VERBATIM)

    add_custom_target(deps_verify
        SOURCES ${PROJECT_SOURCE_DIR}/go.mod
        COMMAND GOPATH=${GO_PATH} GO111MODULE=on GOOS=${CROSS_GOOS} GOARCH=${CROSS_GOARCH} ${GO_EXECUTABLE} mod verify
        WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}"
        COMMENT "Verifying dependencies (go mod verify)"
        VERBATIM)

    add_custom_target(deps_vendor
        SOURCES ${PROJECT_SOURCE_DIR}/go.mod
        COMMAND GOPATH=${GO_PATH} GO111MODULE=on GOOS=${CROSS_GOOS} GOARCH=${CROSS_GOARCH} ${GO_EXECUTABLE} mod vendor
        WORKING_DIRECTORY "${PROJECT_SOURCE_DIR}"
        COMMENT "Creating vendor directory (go mod vendor)"
        VERBATIM)
endif()
