/*
 * Generated by asn1c-0.9.29 (http://lionet.info/asn1c)
 * From ASN.1 module "E2AP-IEs"
 * 	found in "e2ap-v01.01.00.asn1"
 * 	`asn1c -fcompound-names -fincludes-quoted -fno-include-deps -findirect-choice -gen-PER -no-gen-OER -D.`
 */

#ifndef	_E2nodeComponentGNB_CU_UP_ID_H_
#define	_E2nodeComponentGNB_CU_UP_ID_H_


#include "asn_application.h"

/* Including external dependencies */
#include "GNB-CU-UP-ID.h"
#include "constr_SEQUENCE.h"

#ifdef __cplusplus
extern "C" {
#endif

/* E2nodeComponentGNB-CU-UP-ID */
typedef struct E2nodeComponentGNB_CU_UP_ID {
	GNB_CU_UP_ID_t	 gNB_CU_UP_ID;
	/*
	 * This type is extensible,
	 * possible extensions are below.
	 */
	
	/* Context for parsing across buffer boundaries */
	asn_struct_ctx_t _asn_ctx;
} E2nodeComponentGNB_CU_UP_ID_t;

/* Implementation */
extern asn_TYPE_descriptor_t asn_DEF_E2nodeComponentGNB_CU_UP_ID;
extern asn_SEQUENCE_specifics_t asn_SPC_E2nodeComponentGNB_CU_UP_ID_specs_1;
extern asn_TYPE_member_t asn_MBR_E2nodeComponentGNB_CU_UP_ID_1[1];

#ifdef __cplusplus
}
#endif

#endif	/* _E2nodeComponentGNB_CU_UP_ID_H_ */
#include "asn_internal.h"