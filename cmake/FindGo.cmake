# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

# The module defines the following variables:
#   GO_FOUND - true if the Go was found
#   GO_EXECUTABLE - path to the executable
#   GOFMT_EXECUTABLE - path to gofmt executable
#   GO_VERSION - Go version number
#   GO_LANG_VERSION - Go version number without patch, only major and minor versions. Used with "go mod tidy"
#   GO_PLATFORM - i.e. linux
#   GO_ARCH - i.e. amd64
# Example usage:
#   find_package(Go 1.2 REQUIRED)

find_program(GO_EXECUTABLE go PATHS ENV GOROOT GOPATH PATH_SUFFIXES bin)
if(GO_EXECUTABLE)
	execute_process(COMMAND ${GO_EXECUTABLE} version OUTPUT_VARIABLE GO_VERSION_OUTPUT OUTPUT_STRIP_TRAILING_WHITESPACE)
    if(GO_VERSION_OUTPUT MATCHES "go version go([0-9]+\\.[0-9]+\\.?[0-9]*)[a-zA-Z0-9]* ([^/]+)/(.*)")
        set(GO_VERSION ${CMAKE_MATCH_1})
        set(GO_PLATFORM ${CMAKE_MATCH_2})
        set(GO_ARCH ${CMAKE_MATCH_3})
        string(REGEX MATCH "([0-9]+\\.[0-9]+)" GO_LANG_VERSION ${GO_VERSION})
    elseif(GO_VERSION_OUTPUT MATCHES "go version devel .* ([^/]+)/(.*)$")
        set(GO_VERSION "99-devel")
        set(GO_LANG_VERSION, "99-devel")
        set(GO_PLATFORM ${CMAKE_MATCH_1})
        set(GO_ARCH ${CMAKE_MATCH_2})
        message("WARNING: Development version of Go being used, can't determine compatibility.")
    endif()
    get_filename_component(path ${GO_EXECUTABLE} DIRECTORY)
    set(GOFMT_EXECUTABLE ${path}/gofmt${CMAKE_EXECUTABLE_SUFFIX})
endif()
mark_as_advanced(GO_EXECUTABLE)

include(FindPackageHandleStandardArgs)
find_package_handle_standard_args(Go REQUIRED_VARS GO_EXECUTABLE GOFMT_EXECUTABLE GO_PLATFORM GO_ARCH GO_VERSION GO_LANG_VERSION VERSION_VAR GO_VERSION)
