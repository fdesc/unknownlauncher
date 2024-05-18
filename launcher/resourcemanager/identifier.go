package resourcemanager

import "fdesc/unknownlauncher/util/gamepath"

var (
   identifier string
   identifierArch string
   identifierArchOld string
)

func setIdentifier() {
   if gamepath.UserOS == "darwin" {
      identifier = "osx"
      if gamepath.UserArch == "arm64" {
         identifierArch = "arm64"
      }
   } else if gamepath.UserOS == "windows" {
      identifier = "windows"
      if gamepath.UserArch == "amd64" {
         identifierArch = "x64"
         identifierArchOld = "64"
      } else if gamepath.UserArch == "386" {
         identifierArch = "x86"
         identifierArchOld = "32"
      } else if gamepath.UserArch == "arm64" {
         identifierArch = "arm64"
      }
   } else if gamepath.UserOS == "linux" {
      identifier = "linux"
      if gamepath.UserArch == "arm64" {
         identifierArch = "aarch_64"
      } else if gamepath.UserArch == "amd64" {
         identifierArch = "x86_64"
      } else if gamepath.UserArch == "386" {
         identifierArch = "i386"
      }
   }
}
