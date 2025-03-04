# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ARG from_image
FROM ${from_image}

USER root

# Required by php/runtime: libtidy5, libpq5, libxml2, libenchant1c2a,
# libpng16-16, libonig4, libjpeg8, libfreetype6, libxslt1.1, libzip4
RUN apt-get update && apt-get install -y --no-install-recommends \
  libtidy5 \
  libpq5 \
  libxml2 \
  libenchant1c2a \
  libpng16-16 \
  libonig4 \
  libjpeg8 \
  libfreetype6 \
  libxslt1.1 \
  libzip4 \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

USER cnb
