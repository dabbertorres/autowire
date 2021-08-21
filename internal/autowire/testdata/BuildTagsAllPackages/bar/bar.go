// Copyright 2018 The Wire Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//+build !wireinject
//go:build !wireinject

// Package bar includes both wireinject and non-wireinject variants.
package bar

import "github.com/dabbertorres/autowire"

// Set provides an unfriendly user greeting.
var Set = autowire.NewSet(autowire.Value("Bah humbug! This is the wrong variant!"))
