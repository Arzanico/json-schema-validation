package server

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"io"
	"json-schema-validation/lib/tkt"
	"net/http"
)

var schemaText = `{
    "$id": "https://ecosystem.xyz.com/canonical/v2/transmission.schema.json",
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "title": "Transmission",
    "type": "object",
    "version": "2.0.0",
    "properties": {
        "transmissionGUID": {
            "type": "string",
            "description": "Global Unique identifier created by the originator for this instance of the electronic transmission."
        },
        "senderName": {
            "type": "string",
            "description": "The party responsible for sending the transaction. Could be Technology partner or carrier's system"
        },
        "senderNamePlatform": {
            "type": "string",
            "description": "Identifies the sender's system that is providing the data for this data set."
        },
        "receiverName": {
            "type": "string",
            "description": "Name of the company receiving the plan configuration details."
        },
        "creationDateTime": {
            "type": "string",
            "format": "date-time",
            "description": "UTC date and time the transmission was created which follows the ISO 8601 format of YYYY-MM-DDTHH:MM:SS"
        },
        "schemaVersionIdentifier": {
            "type": "string",
            "description": "Identifies the version of the Canonical Model that is being adhered to in this data set."
        },
        "data": {
            "oneOf": [
                {
                    "$ref": "#/$defs/GroupPolicy"
                },
                {
                    "$ref": "#/$defs/RfpQuoting"
                }
            ],
            "description": "The main payload for the transmission. It could be Group Policy/RFP Request/Quote etc."
        },
        "audit": {
            "$ref": "#/$defs/TransmissionAudit",
            "description": "An instance of auditable information provided to ensure accuracy and consistency of the exchanged information."
        }
    },
    "$defs": {
        "Carrier": {
            "type": "object",
            "properties": {
                "identifier": {
                    "type": "string",
                    "description": "This is the predefined internal carrier identifier, set by the Ecosystem and used by the consumer systems when the carrier identification is needed."
                },
                "name": {
                    "type": "string",
                    "description": "This is the carrier name"
                }
            }
        },
        "Contact": {
            "type": "object",
            "description": "Definition of Contact",
            "properties": {
                "fullName": {
                    "type": "string",
                    "description": "Full name used when names are not separated."
                },
                "firstName": {
                    "type": "string",
                    "description": "First name of the contact."
                },
                "middleName": {
                    "type": "string",
                    "description": "Middle name of the contact."
                },
                "lastName": {
                    "type": "string",
                    "description": "Last name of the contact."
                },
                "workPhone": {
                    "type": "string",
                    "description": "Work phone for the contact."
                },
                "workEmail": {
                    "type": "string",
                    "description": "Work email for the contact."
                },
                "isPrimary": {
                    "type": "boolean",
                    "description": "Flag that determines if this is the primary contact."
                },
                "role": {
                    "type": "string",
                    "description": "The role of the contact. Eg. Benefits Admin, HR etc."
                }
            }
        },
        "Address": {
            "properties": {
                "firstLine": {
                    "type": "string",
                    "description": "First line of the address."
                },
                "secondLine": {
                    "type": "string",
                    "description": "Second line of the address."
                },
                "thirdLine": {
                    "type": "string",
                    "description": "Third line of the address."
                },
                "cityName": {
                    "type": "string",
                    "description": "City for the address."
                },
                "stateProvinceCode": {
                    "type": "string",
                    "description": "Code/abbreviation for the state or province for the address."
                },
                "postalCode": {
                    "type": "string",
                    "description": "Postal code for the address."
                },
                "countryCode": {
                    "type": "string",
                    "description": "Country code for the country on the address.",
                    "default": "US"
                },
                "required": [
                    "firstLine",
                    "cityName",
                    "stateProvinceCode",
                    "postalCode"
                ]
            }
        },
        "Location": {
            "properties": {
                "address": {
                    "$ref": "#/$defs/Address",
                    "description": "This is the address of the location."
                }
            }
        },
        "Employer": {
            "properties": {
                "name": {
                    "type": "string",
                    "description": "This is the definition of the Employer element. It represents the company, its contact and address information."
                },
                "federalEmployerIdentificationNumber": {
                    "type": "string",
                    "description": "This is the Employer Identification Number assigned to the employer by IRS."
                },
                "sicCode": {
                    "type": "string",
                    "description": "The national Standard Industrial Classification code associated assigned to the company."
                },
                "contacts": {
                    "type": "array",
                    "description": "This is the list of contacts for the company.",
                    "items": {
                        "$ref": "#/$defs/Contact"
                    }
                },
                "locations": {
                    "type": "array",
                    "description": "This is the list of addresses for the company.",
                    "items": {
                        "$ref": "#/$defs/Location"
                    }
                },
                "required": [
                    "name"
                ]
            }
        },
        "BillGroup": {
            "properties": {
                "identifier": {
                    "type": "string",
                    "description": "This is the identifier fo the Billing Group."
                },
                "name": {
                    "type": "string",
                    "description": "This is the name of the Billing Group."
                },
                "description": {
                    "type": "string",
                    "description": "This property provide description of the Billing Group."
                }
            }
        },
        "BenefitClass": {
            "properties": {
                "identifier": {
                    "type": "string",
                    "description": "This is the identifier of the Benefit Class."
                },
                "name": {
                    "type": "string",
                    "description": "This is the name of the Benefit Class."
                },
                "coverageEffectiveDate": {
                    "type": "string",
                    "description": "This is the effective of coverage for this Benefit Class."
                },
                "minWeeklyEligibleHours": {
                    "type": "string",
                    "description": "Min number of hours an employee must work to be eligible for benefits for this Benefit Class."
                }
            }
        },
        "Benefit": {
            "properties": {
                "name": {
                    "type": "string",
                    "description": "Name of the Benefit. This can be used for display, caption etc."
                },
                "code": {
                    "type": "string",
                    "description": "Code for the Benefit. This is a predefined list of code to uniquely identify a Benefit. Eg. PCP_COPAY_INN for In network copay for doctor's office visit."
                },
                "type": {
                    "type": "string",
                    "description": "this is the type of payment for this service. Eg. COPAY, DEDUCTIBLE, COINSURANCE etc."
                },
                "network": {
                    "type": "string",
                    "description": "This is the network this Benefit value applies to. Eg. In Network, Out of Network."
                },
                "coverageTierCode": {
                    "type": "string",
                    "description": "This is the Tier Code the benefit applies to. Eg. Individual, Family."
                },
                "amount": {
                    "type": "string",
                    "description": "This is the amount provided/covered for the Benefit. Eg. 50% for Coinsurance, $50 for copay."
                },
                "value": {
                    "type": "number",
                    "description": "This is the numeric value for the amount. eg. .5 for 50% coinsurance, 50 for $50 copay."
                },
                "maxAllowed": {
                    "type": "object",
                    "description": "maximum amount allowed for the service over a certain period",
                    "properties": {
                        "amount": {
                            "type": "number",
                            "description": ""
                        },
                        "limit": {
                            "type": "string",
                            "description": "This the limitation in different units like, in duration , in # of visits etc. The unit is described in the unit field. Ex. for 12 months {... limit:12, unit:month ...} and for 4 visits {... limit:4, unit:visit ...}"
                        },
                        "limitUnit": {
                            "type": "string",
                            "description": ""
                        }
                    }
                },
                "frequency": {
                    "type": "string",
                    "description": "Frequency for the service provided for this benefit."
                },
                "category": {
                    "type": "string",
                    "description": "This is a field to group certain benefits into a single category."
                }
            }
        },
        "AgeLimit": {
            "properties": {
                "maxBenefitAge": {
                    "type": "integer",
                    "description": "Upper limit of age at which the benefits are provided."
                },
                "maxChildAge": {
                    "type": "integer",
                    "description": "This is the upper age at which a child dependent can be enrolled."
                },
                "maxStudentAge": {
                    "type": "integer",
                    "description": "This is the upper age at which a full-time student dependent can be enrolled."
                },
                "terminationRule": {
                    "type": "string",
                    "description": "This discloses when the termination will occur. Eg. EventDate, Next Plan Year etc."
                }
            }
        },
        "Rate": {
            "properties": {
                "ageBandLower": {
                    "type": "integer",
                    "description": "Lower age of particular age band"
                },
                "ageBandUpper": {
                    "type": "integer",
                    "description": "Upper age of particular band."
                },
                "coverageTierCode": {
                    "$ref": "#/$defs/CoverageTierEnum",
                    "description": "Indicates the specific tier the RateAmount is pertinent to (Employee, Family, etc)."
                },
                "numberOfLives": {
                    "type": "integer",
                    "description": "This field is the count of lives that matters for this rate. Depending on the Line of Coverage, it can represent different counts."
                },
                "gender": {
                    "type": "string",
                    "description": "Indicates the gender if it is gender specific. May not be applicable to all lines of coverage.",
                    "enum": [
                        "male",
                        "female"
                    ]
                },
                "isTobaccoRated": {
                    "type": "boolean",
                    "description": "Indicates if the rate is specific to tobacco use."
                },
                "rate": {
                    "type": "number",
                    "description": "The number indicating the monthly rate."
                },
                "volume": {
                    "type": "string",
                    "description": "This is the volume to be rated for this category."
                },
                "unit": {
                    "type": "string",
                    "description": "This is the base unit for the amount like, USD/VOLUME PER 100 etc."
                },
                "ratingPeriod": {
                    "type": "string",
                    "description": "This indicates the policy period for the rate.",
                    "default": "current",
                    "enum": [
                        "current",
                        "renewal"
                    ]
                }
            }
        },
        "RateSchedule": {
            "properties": {
                "rateDesign": {
                    "type": "string",
                    "description": "Describes whether the rates are Age-banded, Composite (tiered) or other."
                },
                "rateEffectiveDate": {
                    "type": "string",
                    "description": "The date the rate was effective (often the same as the plan effective date, but not always)."
                },
                "numberOfLives": {
                    "type": "integer",
                    "description": "This field is the count of lives that matters for this rate. Depending on the Line of Coverage, it can represent different counts."
                },
                "pepm": {
                    "type": "number",
                    "description": "Per Employee Per Month. This is the rate used mainly for some ASO.",
                    "format": "decimal"
                },
                "rates": {
                    "type": "array",
                    "description": "Rate details by age/tier/gender/tobacco etc.",
                    "items": {
                        "$ref": "#/$defs/Rate"
                    }
                },
                "roundingRule": {
                    "type": "string",
                    "description": "---- Need to get the description for this field -----"
                },
                "ageBasedOn": {
                    "type": "string",
                    "description": " Whose age to use when calculating premium. ",
                    "enum": [
                        "employee",
                        "insured",
                        "olderOfInsureds",
                        "youngerOfInsureds"
                    ]
                }
            }
        },
        "ContributionTypeEnum": {
            "type": "string",
            "description": "This fields what the contribution type is.",
            "enum": [
                "percentage",
                "flat",
                "contributory",
                "nonContributory",
                "voluntary"
            ]
        },
        "DhmoBenefitDetail": {
            "properties": {
                "adaCode": {
                    "type": "string",
                    "description": "National medical code attached to service name (used in claims/billing)."
                },
                "serviceName": {
                    "type": "string",
                    "description": "Name of service pertinent to the ADA code."
                },
                "memberCost": {
                    "type": "string",
                    "description": "Dollar amount indicating the participant cost for the service name."
                }
            }
        },
        "DentalBenefit": {
            "properties": {
                "networkBenefits": {
                    "type": "array",
                    "description": "This is an array of benefits for each network and outside of network. In most cases, if not always, it should only have 2 items, inNetwork and outOfNetwork elements.",
                    "items": {
                        "network": {
                            "type": "string",
                            "description": "Name of Network pertinent to benefits following.",
                            "example": [
                                "inNetwork",
                                "outOfNetwork"
                            ]
                        },
                        "individualDeductible": {
                            "type": "string",
                            "description": "Dollar amount of the deductible for an individual."
                        },
                        "familyDeductible": {
                            "type": "string",
                            "description": "Dollar amount of the deductible for family."
                        },
                        "preventativeType1": {
                            "type": "string",
                            "description": "Indicates the % of cost the carrier will cover for this service"
                        },
                        "basicType2": {
                            "type": "string",
                            "description": "Indicates the % of cost the carrier will cover for this service."
                        },
                        "majorType3": {
                            "type": "string",
                            "description": "Indicates the % of cost the carrier will cover for this service."
                        },
                        "orthoType4": {
                            "type": "string",
                            "description": "Indicates the % of cost the carrier will cover for this service."
                        },
                        "orthoAgeLimit": {
                            "type": "string",
                            "description": "Indicates the limiting age at which service is no longer covered."
                        },
                        "orthoMaximum": {
                            "type": "string",
                            "description": "Indicates the maximum dollar amount carrier will pay for this service."
                        },
                        "endodontics": {
                            "type": "string",
                            "description": "Indicates the % of cost the carrier will cover for this service."
                        },
                        "periodonticsNonSurgical": {
                            "type": "string",
                            "description": "Indicates the % of cost the carrier will cover for this service."
                        },
                        "periodonticsSurgical": {
                            "type": "string",
                            "description": "Indicates the % of cost the carrier will cover for this service."
                        },
                        "oralSurgerySimple": {
                            "type": "string",
                            "description": "Indicates the % of cost the carrier will cover for this service."
                        },
                        "oralSurgeryComplex": {
                            "type": "string",
                            "description": "Indicates the % of cost the carrier will cover for this service."
                        },
                        "implants": {
                            "type": "string",
                            "description": "Indicates the % of cost the carrier will cover for this service."
                        }
                    }
                },
                "isDeductibleWaivedPreventative": {
                    "type": "boolean",
                    "description": "Indicates whether the deductible will be waived for preventative services."
                },
                "deductibleTimePeriod": {
                    "type": "string",
                    "description": "Indicates the period of time that the deductible applies."
                },
                "annualMaximum": {
                    "type": "string",
                    "description": "Indicates maximum dollar amount carrier will pay annually toward specific service."
                },
                "outOfNetworkReimbursement": {
                    "type": "string",
                    "description": "Indicates amount carrier will reimburse for an out of network claim."
                },
                "maximumRollover": {
                    "type": "string",
                    "description": "Indicates if the carrier offers a rollover option for Annual Max not met."
                },
                "serviceWaitingPeriods": {
                    "type": "string",
                    "description": "Indicates a period of time, if any, a specific service may require the participant to wait before they can utilize that benefit."
                },
                "dhmoBenefits": {
                    "description": "",
                    "type": "array",
                    "items": {
                        "$ref": "#/$defs/DhmoBenefitDetail"
                    }
                }
            }
        },
        "VisionBenefit": {
            "properties": {
                "networkBenefits": {
                    "type": "array",
                    "description": "This is an array of benefits for each network and outside of network. In most cases, if not always, it should only have 2 items, inNetwork and outOfNetwork elements.",
                    "items": {
                        "network": {
                            "type": "string",
                            "description": "Name of Network pertinent to benefits following.",
                            "example": [
                                "inNetwork",
                                "outOfNetwork"
                            ]
                        },
                        "examCopay": {
                            "type": "string",
                            "description": "The amount the insured will pay towards this service."
                        },
                        "materialsCopay": {
                            "type": "string",
                            "description": "The amount the insured will pay towards this service."
                        },
                        "singleLenses": {
                            "type": "string",
                            "description": "Indicates the portion of cost covered by carrier after copay has been met."
                        },
                        "bifocalLenses": {
                            "type": "string",
                            "description": "Indicates the portion of cost covered by carrier after copay has been met."
                        },
                        "trifocalLenses": {
                            "type": "string",
                            "description": "Indicates the portion of cost covered by carrier after copay has been met."
                        },
                        "frames": {
                            "type": "string",
                            "description": "Indicates the portion of cost covered by carrier after copay has been met."
                        },
                        "contactsElective": {
                            "type": "string",
                            "description": "Indicates the portion of cost covered by carrier after copay has been met."
                        },
                        "contactsMedicallyNecessary": {
                            "type": "string",
                            "description": "Indicates the portion of cost covered by carrier after copay has been met."
                        }
                    }
                },
                "examFrequency": {
                    "type": "string",
                    "description": "Indicates the time period within which the benefit for this service will apply once."
                },
                "lensesFrequency": {
                    "type": "string",
                    "description": "Indicates the time period within which the benefit for this service will apply once."
                },
                "framesFrequency": {
                    "type": "string",
                    "description": "Indicates the time period within which the benefit for this service will apply once."
                }
            }
        },
        "StdBenefit": {
            "properties": {
                "benefitAmount": {
                    "type": "string",
                    "description": "Indicates the percentage of salary benefit will pay."
                },
                "maximumAmount": {
                    "type": "string",
                    "description": "Indicates the total max dollar amount benefit will pay."
                },
                "eliminationPeriodAccident": {
                    "type": "string",
                    "description": "Indicates the period of time after accident event that must pass before they begin receiving benefit."
                },
                "eliminationPeriodIllness": {
                    "type": "string",
                    "description": "Indicates the period of time after start of illness that must pass before they begin receiving benefit."
                },
                "benefitDuration": {
                    "type": "string",
                    "description": "Indicates the amount of time the carrier will pay out the benefit."
                },
                "preExistingCondition": {
                    "type": "string",
                    "description": "Indicates the amount of months before eff date AND after effective date participant needs to be disability free to claim benefit."
                }
            }
        },
        "LtdBenefit": {
            "properties": {
                "benefitAmount": {
                    "type": "string",
                    "description": "Indicates the percentage of salary benefit will pay."
                },
                "maximumAmount": {
                    "type": "string",
                    "description": "Indicates the total max dollar amount benefit will pay."
                },
                "definitionOfDisability": {
                    "type": "string",
                    "description": "Indicates carrier defined requirements to be eligible for benefit."
                },
                "gainfulEarningsTest": {
                    "type": "string",
                    "description": "Defines the carrier's required 'loss of earnings' that must be met to be eligible for benefit."
                },
                "eliminationPeriodAccident": {
                    "type": "string",
                    "description": "Indicates the period of time after accident event that must pass before they begin receiving benefit."
                },
                "eliminationPeriodIllness": {
                    "type": "string",
                    "description": "Indicates the period of time after start of illness that must pass before they begin receiving benefit."
                },
                "benefitDuration": {
                    "type": "string",
                    "description": "Indicates the amount of time the carrier will pay out the benefit."
                },
                "specialConditionsLimitations": {
                    "type": "string",
                    "description": "Indicates the time period limitation of benefit duration - conditions that may not have objective medical test - Fibro, Restless leg, chronic fatigue etc."
                },
                "mentalIllness": {
                    "type": "string",
                    "description": "Indicates the time period limitation of benefit duration."
                },
                "substanceAbuse": {
                    "type": "string",
                    "description": "Indicates the time period limitation of benefit duration."
                },
                "preExistingCondition": {
                    "type": "string",
                    "description": "Indicates the amount of months before eff date AND after effective date participant needs to be disability free to claim benefit."
                },
                "rehab": {
                    "type": "string",
                    "description": "Defines the carrier rehabilitation requirement to be eligible for benefit."
                }
            }
        },
        "BasicLifeAdndBenefit": {
            "properties": {
                "benefitAmount": {
                    "type": "string",
                    "description": "Indicates the volume of benefit being quoted."
                },
                "guaranteeIssue": {
                    "type": "string",
                    "description": "Indicates the amount of Life Insurance volume available without having to provide Evidence of Insurability."
                },
                "ageReductionSchedule": {
                    "$ref": "#/$defs/ReductionSchedule",
                    "description": "Indicates the age and associated percentage of benefit reduction being quoted."
                },
                "portability": {
                    "type": "string",
                    "description": "Indicates if policy holder can continue policy after separation of employment and whether Evidence of Insurability would be required.",
                    "enum": [
                        "notIncluded",
                        "withEOI",
                        "withoutEOI"
                    ]
                }
            }
        },
        "VoluntaryLifeAdndBenefit": {
            "properties": {
                "benefitDescription": {
                    "type": "string",
                    "description": "Describes the benefit configuration available."
                },
                "benefitMaximum": {
                    "type": "string",
                    "description": "Indicates the maximum amount of benefit the employee can elect."
                },
                "guaranteeIssue": {
                    "type": "string",
                    "description": "Indicates the amount of Life Insurance volume available without having to provide Evidence of Insurability."
                },
                "dependentAges": {
                    "type": "string",
                    "description": "**** Need to understand the structure better.***** Defines benefit by age range for child life"
                },
                "amountNotToExceed": {
                    "type": "string",
                    "description": "Indicates limitation on Spouse or Child benefit"
                },
                "studentStatus": {
                    "type": "string",
                    "description": "Indicates if carrier offers coverage past 'child' age up to specific age."
                },
                "adndIncluded": {
                    "type": "string",
                    "description": "Indicates if AD&D benefit is included with Life benefit - as option."
                },
                "adndTiedToLifeElection": {
                    "type": "string",
                    "description": "Indicates if AD&D benefit amount must match vol life benefit amount."
                },
                "ageReductionSchedule": {
                    "$ref": "#/$defs/ReductionSchedule",
                    "description": "Indicates the age and associated percentage of benefit reduction being quoted"
                },
                "portability": {
                    "type": "string",
                    "description": "Indicates if policy holder can continue policy after separation of employment and whether Evidence of Insurability would be required.",
                    "enum": [
                        "notIncluded",
                        "withEOI",
                        "withoutEOI"
                    ]
                }
            }
        },
        "AccidentBenefit": {
            "properties": {
                "hoursCovered": {
                    "type": "string",
                    "description": "Indicates when the accident event may occur to be eligible for coverage."
                },
                "hospitalAdmission": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "icuAdmission": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "dailyHospitalConfinement": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "dailyIcuConfinement": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "ambulanceAir": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "ambulanceGround": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "emergencyTreatment": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "urgentCare": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "initialPhysicianOfficeVisit": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "dislocations": {
                    "type": "string",
                    "description": "Max benefit amount."
                },
                "fractures": {
                    "type": "string",
                    "description": "Max benefit amount."
                },
                "enhancedBenefitOrganizedSportsRelatedAccident": {
                    "type": "string",
                    "description": "Max benefit amount."
                },
                "portability": {
                    "type": "string",
                    "description": "Indicates if policy holder can continue policy after separation of employment and whether Evidence of Insurability would be required."
                },
                "wellness": {
                    "type": "string",
                    "description": "Max benefit amount."
                },
                "ageReductions": {
                    "$ref": "#/$defs/ReductionSchedule",
                    "description": "Indicates the age and associated percentage of benefit reduction being quoted."
                }
            }
        },
        "CriticalIllnessBenefit": {
            "properties": {
                "benefitDescription": {
                    "type": "string",
                    "description": "Describes the benefit configuration available."
                },
                "benefitMaximum": {
                    "type": "string",
                    "description": "Indicates the maximum amount of benefit the employee can elect."
                },
                "guaranteeIssue": {
                    "type": "string",
                    "description": "Indicates the amount of Life Insurance volume available without having to provide Evidence of Insurability."
                },
                "invasiveMalignantCancer": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "category1Vascular": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "organKidneyFailure": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "maximumPayout": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "occurrenceOfDiffIllness": {
                    "type": "string",
                    "description": "Indicates the percentage at which an additional illness will be covered, and the time period compared to the initial illness covered."
                },
                "additionalOccurrenceOfSameIllness": {
                    "type": "string",
                    "description": "Indicates the percentage at which a repeat occurrence of the same illness will be covered, and the time period compared to the initial illness covered."
                },
                "ageReduction": {
                    "$ref": "#/$defs/ReductionSchedule",
                    "description": "Details any benefit change-increments based on age."
                },
                "preExistingCondition": {
                    "type": "string",
                    "description": "Indicates the amount of months before eff date AND after effective date participant needs to be illness free to claim benefit."
                },
                "portability": {
                    "type": "string",
                    "description": "Indicates if policy holder can continue policy after separation of employment and whether Evidence of Insurability would be required."
                },
                "wellness": {
                    "type": "string",
                    "description": "Indicates dollar amount of available reimbursement for member completing wellness initiative(s)."
                },
                "ageBasis": {
                    "type": "string",
                    "description": "Indicates whether the available benefit(s) are based on member age at time of policy issuance, or based member attained age in conjunction with age-band rules.",
                    "enum": [
                        "attainedAge",
                        "issueAge"
                    ]
                }
            }
        },
        "CancerBenefit": {
            "properties": {
                "diagnosisBenefit": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "initialDiagnosisWaitingPeriod": {
                    "type": "string",
                    "description": "Indicates time period member must wait after initial diagnosis before benefit applies."
                },
                "radiationTherapyChemo": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "bloodPlasmaPlatelets": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "hospice": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "hospitalConfinement": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "icuConfinement": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "skinCancer": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "surgicalBenefit": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "preExistingCondition": {
                    "type": "string",
                    "description": "Indicates the amount of months before eff date AND after effective date participant needs to be cancer free to claim benefit."
                },
                "portability": {
                    "type": "string",
                    "description": "Indicates if policy holder can continue policy after separation of employment and whether Evidence of Insurability would be required."
                },
                "wellness": {
                    "type": "string",
                    "description": "Indicates dollar amount of available reimbursement for member completing wellness initiative(s)."
                },
                "waiverOfPremium": {
                    "type": "string",
                    "description": "Indicates if premiums are waived and for what time period, should the member become disabled."
                }
            }
        },
        "HospitalIndemnityBenefit": {
            "properties": {
                "hospitalIcuAdmission": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "dailyHospitalIcuConfinement": {
                    "type": "string",
                    "description": "Indicates dollar amount carrier will cover as benefit for this service, including any conditional requirements."
                },
                "preExistingCondition": {
                    "type": "string",
                    "description": "Indicates the amount of months before eff date AND after effective date participant needs to be NOT ADMITTED in hospital to claim benefit (waiting period essentially)."
                },
                "portability": {
                    "type": "string",
                    "description": "Indicates if policy holder can continue policy after separation of employment and whether Evidence of Insurability would be required."
                },
                "wellness": {
                    "type": "string",
                    "description": "Indicates dollar amount of available reimbursement for member completing wellness initiative(s)."
                }
            }
        },
        "FmlaBenefit": {
            "properties": {
                "federalFmla": {
                    "type": "string",
                    "description": "federalFmla"
                },
                "stateLeaves": {
                    "type": "string",
                    "description": "stateLeaves"
                },
                "militaryUserra": {
                    "type": "string",
                    "description": "militaryUSERRA"
                },
                "juryDuty": {
                    "type": "string",
                    "description": "juryDuty"
                },
                "ada": {
                    "type": "string",
                    "description": "ADA"
                },
                "historyAndTakeover": {
                    "type": "string",
                    "description": "historyAndTakeover"
                },
                "companyLeaves": {
                    "type": "string",
                    "description": "companyLeaves"
                },
                "correspondence": {
                    "type": "string",
                    "description": "correspondence"
                },
                "integratedStdFmlaClaimIntake": {
                    "type": "string",
                    "description": "integratedStdFmlaClaimIntake"
                }
            }
        },
        "EapBenefit": {
            "properties": {
                "faceToFaceVisits": {
                    "type": "string",
                    "description": "Indicates the number of mental health visits allowed per year."
                },
                "perOccurrencePerYear": {
                    "type": "string",
                    "description": "Indicates the number of allowed occurrences, per time period."
                },
                "unlimitedTelephonic": {
                    "type": "string",
                    "description": "Indicates if EAP includes telephonic provider services."
                },
                "legalFinancialResources": {
                    "type": "string",
                    "description": "Indicates if EAP includes Legal and Financial services."
                },
                "additionalResourcesIncluded": {
                    "type": "string",
                    "description": "Indicates additional benefits included in the plan."
                },
                "tiedToAnotherLoc": {
                    "type": "string",
                    "description": "Indicates which, if any, LoC the EAP plan is tied to."
                }
            }
        },
        "IndividualDisabilityBenefit": {
            "properties": {
                "ltdPlanDesign": {
                    "type": "string",
                    "description": ""
                },
                "idiPlanDesign": {
                    "type": "string",
                    "description": ""
                },
                "gsiBenefitMaximum": {
                    "type": "string",
                    "description": ""
                },
                "definitionOfDisability": {
                    "type": "string",
                    "description": "Indicates carrier defined requirements to be eligible for benefit."
                },
                "eliminationPeriod": {
                    "type": "string",
                    "description": "Indicates the period of time after participant becomes disabled that must pass before they begin receiving benefit."
                },
                "benefitPeriod": {
                    "type": "string",
                    "description": "Indicates the time period limitation of plan and benefits."
                },
                "preExistingCondition": {
                    "type": "string",
                    "description": "Indicates the amount of months before eff date AND after effective date participant needs to be disability free to claim benefit."
                },
                "mentalDisorderBenefit": {
                    "type": "string",
                    "description": "Indicates the time period limitation of benefit duration."
                },
                "recoveryBenefit": {
                    "type": "string",
                    "description": "Indicates the time period limitation of benefit duration."
                },
                "catastrophicBenefit": {
                    "type": "string",
                    "description": "Indicates the time period limitation of benefit duration."
                },
                "portability": {
                    "type": "string",
                    "description": "Indicates if policy holder can continue policy after separation of employment and whether Evidence of Insurability would be required."
                }
            }
        },
        "WholeLifeBenefit": {
            "properties": {
                "benefitAmount": {
                    "type": "string",
                    "description": "Describes the benefit configuration available."
                },
                "guaranteeIssue": {
                    "type": "string",
                    "description": "Indicates the amount of benefit available without having to provide Evidence of Insurability."
                },
                "ridersIncluded": {
                    "type": "array",
                    "description": "Indicates any riders included with whole life.",
                    "items": {
                        "type": "string",
                        "description": "Each rider described as an item in the array."
                    }
                },
                "additionalOptionsIncluded": {
                    "type": "string",
                    "description": "Indicates any additional benefit options or provisions."
                },
                "portability": {
                    "type": "string",
                    "description": "Indicates if policy holder can continue policy after separation of employment and whether Evidence of Insurability would be required."
                },
                "interestRate": {
                    "type": "string",
                    "description": "interestRate"
                },
                "endows": {
                    "type": "string",
                    "description": "endows"
                },
                "serviceWaitingPeriod": {
                    "type": "string",
                    "description": ""
                }
            }
        },
        "CoverageTierEnum": {
            "type": "string",
            "description": "Values acceptable for coverage tier.",
            "enum": [
                "employee",
                "employeeDependent",
                "employeeSpouse",
                "employeeChildren",
                "employeeFamily",
                "employee2Dependents",
                "spouseOnly",
                "spouseDependent",
                "spouseChildren",
                "childOnly"
            ]
        },
        "BenefitPlan": {
            "properties": {
                "identifier": {
                    "type": "string",
                    "description": "This is the plan identifier."
                },
                "hiosIdentifier": {
                    "type": "string",
                    "description": "This is the HIOS identifier. This could be potentially same as the identifier."
                },
                "carrier": {
                    "$ref": "#/$defs/Carrier",
                    "description": "Indicates the Carrier of record for the current plan."
                },
                "name": {
                    "type": "string",
                    "description": "Name of the Benefit Plan. This is the name used in booklets and other published materials."
                },
                "type": {
                    "type": "string",
                    "description": "Universal identifier for a specific logical product group that may be elected by a member. Usually describes basic plan design (PPO, HMO, Basic, Voluntary, etc)."
                },
                "effectiveDate": {
                    "type": "string",
                    "description": "Date the plan coverage initially begins. The date must be in the format of YYYY-MM-DD."
                },
                "policyPeriod": {
                    "type": "string",
                    "description": "This field tells us if the plan is for current policy period or renewal policy period. For renewal RFP, this should be 'alternate' if different from current. For new business, use newBusiness",
                    "enum": [
                        "newBusiness",
                        "current",
                        "alternate",
                        "renewal"
                    ]
                },
                "stateCode": {
                    "type": "string",
                    "description": "This is the state code from the plan. It can be different from the group's legal address."
                },
                "fundingType": {
                    "type": "string",
                    "description": "This is the funding type for the coverage. it represents values like FULLY_INSURED, ADMINISTRATIVE_SERVICE_ONLY etc."
                },
                "erisaStatus": {
                    "type": "string",
                    "description": ""
                },
                "cobraEligible": {
                    "type": "string",
                    "description": "Indicates the coverage is being continued after leaving employment under the Consolidated Omnibus Budget Reconciliation Act of 1985 (<COBRA)."
                },
                "rateGuarantee": {
                    "type": "string",
                    "description": "---- Need description ----"
                },
                "participationRequirement": {
                    "type": "string",
                    "description": "---- Need description ----"
                },
                "rateCaps": {
                    "type": "string",
                    "description": "indicates if there is a cap to rate increases for any specific time *will apply to return quote only"
                },
                "tierOptions": {
                    "$ref": "#/$defs/CoverageTierEnum",
                    "description": "This is the coverage tier for the plan. LDEx CoverageTier"
                },
                "networkProviderName": {
                    "type": "string",
                    "description": "This is used mainly for Vision to indicate which network is used. Eg. VSP, EysMed etc.",
                    "example": [
                        "VSP",
                        "EyeMed"
                    ]
                },
                "optionName": {
                    "type": "string",
                    "description": "This is used mainly for quotes with multiple plan options. This is when there are different plan options even for the same benefit class.",
                    "example": [
                        "High",
                        "Low"
                    ]
                },
                "benefitClassAvailability": {
                    "type": "array",
                    "description": "Indicates which BenefitClassIdentifier are eligible for this plan (0001, etc - may list more than one).",
                    "items": {
                        "type": "string"
                    }
                },
                "rateSchedule": {
                    "type": "array",
                    "description": "This is the rate schedule for this plan.",
                    "items":{
                        "$ref": "#/$defs/RateSchedule"
                    }
                },
                "benefits": {
                    "oneOf": [
                        {
                            "$ref": "#/$defs/DentalBenefit"
                        },
                        {
                            "$ref": "#/$defs/VisionBenefit"
                        },
                        {
                            "$ref": "#/$defs/BasicLifeAdndBenefit"
                        },
                        {
                            "$ref": "#/$defs/VoluntaryLifeAdndBenefit"
                        },
                        {
                            "$ref": "#/$defs/StdBenefit"
                        },
                        {
                            "$ref": "#/$defs/LtdBenefit"
                        },
                        {
                            "$ref": "#/$defs/AccidentBenefit"
                        },
                        {
                            "$ref": "#/$defs/CriticalIllnessBenefit"
                        },
                        {
                            "$ref": "#/$defs/HospitalIndemnityBenefit"
                        },
                        {
                            "$ref": "#/$defs/FmlaBenefit"
                        },
                        {
                            "$ref": "#/$defs/EapBenefit"
                        },
                        {
                            "$ref": "#/$defs/IndividualDisabilityBenefit"
                        },
                        {
                            "$ref": "#/$defs/WholeLifeBenefit"
                        }
                    ]
                }
            }
        },
        "WaitingPeriodRule": {
            "properties": {
                "employeeType": {
                    "type": "string",
                    "description": "This is the type of employee the rule applies to. Example, new, rehire, current, future, special etc."
                },
                "periodType": {
                    "type": "string",
                    "description": "This is the type of rule to be applied. For example, hireDate. This can also contain a variable part that is replaced with the amount and unit to determine the rule. For example, firstOfMonthAfterX where X should be replaced by the amount and unit. So, if the amount is 15 and unit is days, it should be first of the month after 15 days."
                },
                "amount": {
                    "type": "integer",
                    "description": "This is the waiting period, to be applied to the period type."
                },
                "unit": {
                    "type": "string",
                    "description": "This is the unit for the amount to be applied to the period type."
                }
            }
        },
        "ReductionDetails": {
            "properties": {
                "ageBegins": {
                    "type": "integer",
                    "description": ""
                },
                "ageEnds": {
                    "type": "integer",
                    "description": ""
                },
                "percentageAmount": {
                    "type": "number",
                    "description": ""
                },
                "rateAmount": {
                    "type": "number",
                    "description": ""
                }
            }
        },
        "ReductionSchedule": {
            "type": "array",
            "description": "This is the list of reductions",
            "items": {
                "$ref": "#/$defs/ReductionDetails"
            }
        },
        "CoverageCompensationSplit": {
            "properties": {
                "type": {
                    "type": "string",
                    "description": " Type of compensation paid to a Producer as either a flat amount or percentage of premium. ",
                    "enum": [
                        "flat",
                        "percentage"
                    ]
                },
                "currentAmount": {
                    "type": "number",
                    "description": " for type:flat, Flat rate that is retained and allocated to Producer in the form of commission. for type:percentage,  Premium percentage that is retained and allocated to producers in the form of commission. ",
                    "format": "decimal"
                },
                "requestedAmount": {
                    "type": "number",
                    "description": " for RFP, this is the requested compensation split value. ",
                    "format": "decimal"
                },
                "benefitPlanIdentifier": {
                    "type": "string",
                    "description": "Value used to reference the Coverage.BenefitPlan.identifier element within the payload."
                }
            }            
        },
        "Coverage": {
            "properties": {
                "identifier": {
                    "type": "string",
                    "description": ""
                },
                "type": {
                    "type": "string",
                    "description": "Dental, Vision, LTD, STD etc.",
                    "enum": [
                        "medical",
                        "dental",
                        "vision",
                        "std",
                        "ltd",
                        "basicLife",
                        "voluntaryLife",
                        "accident",
                        "criticalIllness",
                        "cancer",
                        "hospitalIndemnity",
                        "fmla",
                        "eap",
                        "individualDisability",
                        "wholeLife"
                    ]
                },
                "numberOfEligibleEmployees": {
                    "type": "integer",
                    "description": "Number of employees eligible to enroll for this type of coverage"
                },
                "terminationDate": {
                    "type": "string",
                    "description": "The last date that the plan is in effect. The date must be in the format of YYYY-MM-DD."
                },
                "terminationRule": {
                    "type": "string",
                    "description": ""
                },
                "employerContributionType": {
                    "$ref": "#/$defs/ContributionTypeEnum",
                    "description": "Indicates the type of Employer contribution "
                },
                "employerContributionAmount": {
                    "type": "number",
                    "description": "Indicates the amount for Employer contribution "
                },
                "requiresEoi": {
                    "type": "boolean",
                    "description": ""
                },
                "benefitPlans": {
                    "type": "array",
                    "description": "",
                    "items": {
                        "$ref": "#/$defs/BenefitPlan"
                    }
                },
                "waitingPeriodRules": {
                    "type": "array",
                    "description": "",
                    "items": {
                        "$ref": "#/$defs/WaitingPeriodRule"
                    }
                }
            }
        },
        "Producer": {
            "description": "Person or an organization that assists in the buying of or enrolling in insurance.",
            "properties": {
                "carrierProducerNumber": {
                    "type": "string",
                    "description": "Carrier assigned identifiers associated with the person or an organization that assists in the buying of or enrolling in insurance."
                },
                "type": {
                    "type": "string",
                    "description": "Type of Producer that assists in the buying of or enrolling in insurance such as the Broker or Servicing Agent.",
                    "enum": [
                        "brokerOfRecord",
                        "servicingAgent",
                        "writingAgent",
                        "other"
                    ]
                },
                "contact": {
                    "description": "Contact details for a person representing the organization that assists in the buying of or enrolling in insurance.",
                    "$ref": "#/$defs/Contact"
                },
                "brokerName": {
                    "description": "Broker/Firm name for the producer",
                    "$ref": "#/$defs/Contact"
                },
                "accountManager": {
                    "description": "contact of the account manger for the producer.",
                    "$ref": "#/$defs/Contact"
                },
                "coverageCompensationSplit": {
                    "type": "array",
                    "description": "Premium percentage that is retained and allocated to Producers in the form of commission.",
                    "items": {
                        "$ref": "#/$defs/CoverageCompensationSplit"
                    }
                }
            }
        },
        "Discount": {
            "properties":{
                "name": {
                    "type": "string",
                    "description": "Name of the discount. Eg. Technology, Autopay etc."
                },
                "detail": {
                    "type": "string",
                    "description": "Use this field for a full description of the discount. It is possible that some of these discounts are already built in. In such case, we won't have the amount and use this field for the information about it."
                },
                "amount": {
                    "type": "string",
                    "description": "Discount amount. Could be flat dollar amount. It could be a percentage."
                },
                "amountType": {
                    "type": "string",
                    "description": "Indicates if this is a flat dollar amount or percentage.",
                    "enum": [
                        "flat",
                        "percentage"
                    ]
                }
            }
        },
        "GroupConfiguration": {
            "properties": {
                "benefitClasses": {
                    "type": "array",
                    "description": "",
                    "items": {
                        "$ref": "#/$defs/BenefitClass"
                    }
                },
                "coverages": {
                    "type": "array",
                    "description": "",
                    "items": {
                        "$ref": "#/$defs/Coverage"
                    }
                },
                "producers": {
                    "type": "array",
                    "description": "",
                    "items": {
                        "$ref": "#/$defs/Producer"
                    }
                },
                "generalAgent": {
                    "type": "string",
                    "description": "Name of the General Agent acting on behalf of the Broker."
                },
                "numberOfEligibleEmployees": {
                    "type": "integer",
                    "description": "Number of employees eligible to enroll "
                },
                "discounts": {
                    "type": "array",
                    "description": "A list of all the discounts/credits.",
                    "items": {
                        "$ref": "#/$defs/Discount"
                    }
                }
            }
        },
        "GroupPolicy": {
            "properties": {
                "masterAgreementNumber": {
                    "type": "string",
                    "description": "Carrier assigned group/contract number. All individual policies or certificates at the group level would roll up to this master group number. This may be the same as the agreement (aka policy) number for some carriers."
                },
                "effectiveDate": {
                    "type": "string",
                    "format": "date",
                    "description": "The date the original policy went into effect between carrier and employer (company). The date must be in the format of YYYY-MM-DD"
                },
                "terminationDate": {
                    "type": "string",
                    "format": "date",
                    "description": "The date the policy terminated/will terminate, if available. The date must be in the format of YYYY-MM-DD"
                },
                "terminationReason": {
                    "type": "string",
                    "description": "Describes why the group policy plan was terminated."
                },
                "status": {
                    "type": "string",
                    "description": "Indicates if a group is active, termed for nonpayment, etc."
                },
                "carrier": {
                    "$ref": "#/$defs/Carrier",
                    "description": "This is the carrier that issued this policy."
                },
                "employer": {
                    "$ref": "#/$defs/Employer",
                    "description": "This is the employer that has been issued this policy."
                },
                "billGroups": {
                    "description": "List of the billing groups",
                    "type": "array",
                    "items": {
                        "$ref": "#/$defs/BillGroup"
                    }
                },
                "groupPolicyConfiguration": {
                    "$ref": "#/$defs/GroupConfiguration",
                    "description": "This is the policy object, contract between the employer and the carrier."
                }
            }
        },
        "RfpQuoting": {
            "properties": {
                "identifier": {
                    "type": "string",
                    "format": "uuid",
                    "description": "This is the identifier of the Request for Proposal. This id is also used by the quotes generated for the RFP, as the foreign key."
                },
                "carrierIssuedIdentifier": {
                    "type": "string",
                    "description": "This is an identifier of the Request for Proposal, issued by the system of record at the carrier. Quotes generated by different carriers for the same request to a specific carrier should share the same value for this RFP identifier."
                },
                "effectiveDate": {
                    "type": "string",
                    "format": "date",
                    "description": "Date when the requested coverage become effective. The date must be in the format of YYYY-MM-DD"
                },
                "dueDate": {
                    "type": "string",
                    "format": "date",
                    "description": "Date by when the requested quotes are due. The date must be in the format of YYYY-MM-DD"
                },
                "notes": {
                    "type": "string",
                    "description": "Notes to be accompanied with the request for things like expedited processing or any extra information."
                },
                "status": {
                    "type": "string",
                    "description": "status",
                    "enum": [
                        "pending",
                        "quoted"
                    ]
                },
                "employer": {
                    "$ref": "#/$defs/Employer",
                    "description": "This is the employer that has requested a proposal."
                },
                "groupConfiguration": {
                    "description": "This represents the current or desired configuration of the employer group.",
                    "$ref": "#/$defs/GroupConfiguration"
                }
            }
        },
        "TransmissionAudit": {
            "properties": {
                "identifier": {
                    "type": "string",
                    "description": "Unique identifier created by the originator for this instance of the electronic transmission."
                },
                "dataType": {
                    "type": "string",
                    "description": "The specific type for the object sent in the property 'data'."
                },
                "additionalProperties": {
                    "type": "integer",
                    "description": "A key value pair <string, integer> to get the counts for different elements in the data."
                }
            }
        }
    }
}
`

func NewHttpServer(addr string) *http.Server {
	httpsrv := newHttpServer()
	r := mux.NewRouter()
	r.HandleFunc("/validate", httpsrv.validate).Methods(http.MethodPost)
	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

type httpServer struct {
	Payload *PayloadValidationRequest
}

func newHttpServer() *httpServer {
	return &httpServer{
		Payload: NewPayloadValidationRequest(),
	}
}

func (s *httpServer) validate(w http.ResponseWriter, r *http.Request) {
	var m interface{}
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(requestBody, &m)
	if err != nil {
		panic(err)
	}

	schema, err := jsonschema.CompileString("", schemaText)

	if err != nil {
		tkt.JsonResponse(err, w)
		return
	}

	err = schema.Validate(m)
	if err != nil {
		tkt.JsonResponse(err, w)
		return
	}

	tkt.JsonResponse(err, w)
	return
}
