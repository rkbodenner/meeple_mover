/**
 * This is used by TransactionLib and CollectorClientLib
 */
#ifndef NEWRELIC_COMMON_H_
#define NEWRELIC_COMMON_H_

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif /* __cplusplus */

/**
 * Return codes
 */
static const int NEWRELIC_RETURN_CODE_OK = 0;
static const int NEWRELIC_RETURN_CODE_OTHER = -0x10001;
static const int NEWRELIC_RETURN_CODE_DISABLED = -0x20001;
static const int NEWRELIC_RETURN_CODE_INVALID_PARAM = -0x30001;
static const int NEWRELIC_RETURN_CODE_INVALID_ID = -0x30002;
static const int NEWRELIC_RETURN_CODE_TRANSACTION_NOT_STARTED = -0x40001;
static const int NEWRELIC_RETURN_CODE_TRANSACTION_IN_PROGRESS = -0x40002;
static const int NEWRELIC_RETURN_CODE_TRANSACTION_NOT_NAMED = -0x40003;

#ifdef __cplusplus
} //! extern "C"
#endif /* __cplusplus */

#endif /* NEWRELIC_COMMON_H_ */
