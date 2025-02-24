//  Copyright (c) 2017-2018 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#ifndef QUERY_UTILS_HPP_
#define QUERY_UTILS_HPP_
#include <cuda_runtime.h>
#include <thrust/pair.h>
#include <thrust/tuple.h>
#include <algorithm>
#include <cfloat>
#include <cmath>
#include <cstdint>
#include <exception>
#include <type_traits>
#include <stdexcept>
#include <string>
#include <utility>
#include "query/time_series_aggregate.h"
#ifdef USE_RMM
#include <rmm/thrust_rmm_allocator.h>
#endif

// We need this macro to define functions that can only be called in host
// mode or device mode, but not both. The reason to have this mode is because
// a "device and host" function can only call "device and host" function. They
// cannot call device-only functions like "atomicAdd" even we call them under
// RUN_ON_DEVICE macro.
#ifdef RUN_ON_DEVICE
#define __host_or_device__ __device__
#else
#define __host_or_device__ __host__
#endif

// This macro is for setting the correct thrust execution policy given whether
// RUN_ON_DEVICE and USE_RMM
#ifdef RUN_ON_DEVICE
#  ifdef USE_RMM
#    define GET_EXECUTION_POLICY(cudaStream) \
       rmm::exec_policy(cudaStream)->on(cudaStream)
#  else
#    define GET_EXECUTION_POLICY(cudaStream) \
        thrust::cuda::par.on(cudaStream)
#  endif
#else
#  define GET_EXECUTION_POLICY(cudaStream) thrust::host
#endif


// This function will check the cuda error of current thread and throw an
// exception if any.
void CheckCUDAError(const char *message);

// AlgorithmError represents a exception class that contains a error message.
class AlgorithmError : public std::exception {
 protected:
  std::string message_;
 public:
  explicit AlgorithmError(const std::string &message);
  virtual const char *what() const throw();
};

namespace ares {

// Parameters for custom kernel.
const unsigned int WARP_SIZE = 32;
const unsigned int STEP_SIZE = 64;
const unsigned int BLOCK_SIZE = 512;

// common_type determines the common type between type A and B,
// that is the type both types can be implicitly converted to.
template <typename A, typename B>
struct common_type {
  typedef typename std::conditional<
      std::is_floating_point<A>::value || std::is_floating_point<B>::value,
      float_t,
      typename std::conditional<
          std::is_same<A, int64_t>::value || std::is_same<B, int64_t>::value,
          int64_t,
          typename std::conditional<std::is_signed<A>::value ||
                                        std::is_signed<B>::value,
                                    int32_t, uint32_t>::type>::type>::type type;
};

// Special common_type for GeoPointT
template<>
struct common_type<GeoPointT, GeoPointT> {
  typedef GeoPointT type;
};

// get_identity_value returns the identity value for the aggregation function.
// Identity value is a special type of element of a set with respect to a
// binary operation on that set, which leaves other elements unchanged when
// combined with them.
template <typename Value>
__host__ __device__ Value get_identity_value(AggregateFunction aggFunc) {
  switch (aggFunc) {
    case AGGR_AVG_FLOAT:return 0;  // zero avg and zero count.
    case AGGR_SUM_UNSIGNED:
    case AGGR_SUM_SIGNED:
    case AGGR_SUM_FLOAT:return 0;
    case AGGR_MIN_UNSIGNED:return UINT32_MAX;
    case AGGR_MIN_SIGNED:return INT32_MAX;
    case AGGR_MIN_FLOAT:return FLT_MAX;
    case AGGR_MAX_UNSIGNED:return 0;
    case AGGR_MAX_SIGNED:return INT32_MIN;
    case AGGR_MAX_FLOAT:return FLT_MIN;
    default:return 0;
  }
}

inline uint8_t getStepInBytes(DataType dataType) {
  switch (dataType) {
    case Bool:
    case Int8:
    case Uint8:return 1;
    case Int16:
    case Uint16:return 2;
    case Int32:
    case Uint32:
    case Float32:return 4;
    case GeoPoint:
    case Int64:
    case Uint64: return 8;
    case UUID: return 16;
    default:
      throw std::invalid_argument(
          "Unsupported data type for VectorPartyInput");
  }
}

inline
__host__ __device__
void setDimValue(uint8_t *outPtr, uint8_t *inPtr, uint16_t dimBytes) {
  switch (dimBytes) {
      case 16:
        *reinterpret_cast<UUIDT *>(outPtr) = *reinterpret_cast<UUIDT *>(inPtr);
      case 8:
        *reinterpret_cast<uint64_t *>(outPtr) =
            *reinterpret_cast<uint64_t *>(inPtr);
      case 4:
        *reinterpret_cast<uint32_t *>(outPtr) =
            *reinterpret_cast<uint32_t *>(inPtr);
      case 2:
        *reinterpret_cast<uint16_t *>(outPtr) =
            *reinterpret_cast<uint16_t *>(inPtr);
      case 1:*outPtr = *inPtr;
    }
}

template<typename kernel>
void calculateDim3(int *grid_size, int *block_size, size_t size, kernel k) {
  int min_grid_size;
  cudaOccupancyMaxPotentialBlockSize(&min_grid_size, block_size, k);
  CheckCUDAError("cudaOccupancyMaxPotentialBlockSize");
  // find needed gridsize
  size_t needed_grid_size = (size + *block_size - 1) / *block_size;
  *grid_size = static_cast<int>(std::min(static_cast<size_t>(min_grid_size),
                                         needed_grid_size));
}

// Set of atomicAdd operator wrappers.
// In device mode, they will call cuda atomicX.
// In host mode, they will just do the addition without atomicity guarantee as
// std::atomic protects on memory managed by itself instead of on a passed-in
// address. This is ok for now since for host mode algorithms are not running
// in parallel.
// TODO(lucafuji): find atomic libraries on host.
#ifdef RUN_ON_DEVICE
template <typename val_type>
__host__ __device__
inline val_type atomicAdd(val_type* address, val_type val) {
  return ::atomicAdd(address, val);
}

#else
template <typename val_type>
__host__ __device__
inline val_type atomicAdd(val_type* address, val_type val) {
  val_type old = *address;
  *address += val;
  return old;
}
#endif


// GPU memory access has to be aligned to 1, 2, 4, 8, 16 bytes
// http://docs.nvidia.com/cuda/cuda-c-programming-guide/index.html#device-memory-accesses
// therefore we do byte to byte comparison here
inline __host__ __device__ bool memequal(const uint8_t *lhs, const uint8_t *rhs,
                                         int bytes) {
  for (int i = 0; i < bytes; i++) {
    if (lhs[i] != rhs[i]) {
      return false;
    }
  }
  return true;
}

template<typename t1, typename t2>
__host__ __device__
thrust::pair<t1, t2> tuple2pair(thrust::tuple<t1, t2> t) {
  return thrust::make_pair(thrust::get<0>(t), thrust::get<1>(t));
}

__host__ __device__ uint32_t murmur3sum32(const uint8_t *key, int bytes,
                                          uint32_t seed);
__host__ __device__ void murmur3sum128(const uint8_t *key, int len,
                                       uint32_t seed, uint64_t *out);

template<int hash_bytes = 64>
struct hash_output_type { using type = uint64_t; };

template<>
struct hash_output_type<32> { using type = uint32_t; };

template<int hash_bytes = 64>
__host__ __device__
inline
typename hash_output_type<hash_bytes>::type
murmur3sum(const uint8_t *key, int bytes, uint32_t seed) {
  uint64_t hashedOutput[2];
  murmur3sum128(key, bytes, seed, hashedOutput);
  return hashedOutput[0];
}

template<>
__host__ __device__
inline
typename hash_output_type<32>::type
murmur3sum<32>(const uint8_t *key, int bytes, uint32_t seed) {
  return murmur3sum32(key, bytes, seed);
}

}  // namespace ares
#endif  // QUERY_UTILS_HPP_
