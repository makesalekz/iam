## OneTimePassword:

|      Field      |           Type           | Unique | Optional | Nillable | Default | UpdateDefault | Immutable |            StructTag             | Validators | Comment |
|---|---|---|---|---|---|---|---|---|---|---|
| id              | int                      | false  | false    | false    | false   | false         | false     | json:"id,omitempty"              |          0 |         |
| user_id         | int64                    | false  | false    | false    | false   | false         | false     | json:"user_id,omitempty"         |          0 |         |
| code            | string                   | false  | false    | false    | false   | false         | false     | json:"code,omitempty"            |          1 |         |
| type            | enum.OneTimePasswordType | false  | false    | false    | false   | false         | false     | json:"type,omitempty"            |          0 |         |
| is_used         | bool                     | false  | false    | false    | true    | false         | false     | json:"is_used,omitempty"         |          0 |         |
| expires_at      | time.Time                | false  | false    | false    | false   | false         | false     | json:"expires_at,omitempty"      |          0 |         |
| created_at      | time.Time                | false  | false    | false    | true    | false         | false     | json:"created_at,omitempty"      |          0 |         |
| failed_attempts | int64                    | false  | false    | false    | true    | false         | false     | json:"failed_attempts,omitempty" |          0 |         |


| Edge | Type | Inverse | BackRef | Relation | Unique | Optional | Comment |
|---|---|---|---|---|---|---|---|
| user | User | false   |         | M2O      | true   | false    |         |

## User:

|       Field       |   Type    | Unique | Optional | Nillable | Default | UpdateDefault | Immutable |             StructTag              | Validators |            Comment             |
|---|---|---|---|---|---|---|---|---|---|---|
| id                | int64     | false  | false    | false    | false   | false         | true      | json:"id,omitempty"                |          0 |                                |
| deleted_at        | time.Time | false  | true     | true     | false   | false         | false     | json:"deleted_at,omitempty"        |          0 |                                |
| remove_at         | time.Time | false  | true     | true     | false   | false         | false     | json:"remove_at,omitempty"         |          0 |                                |
| phone             | string    | true   | true     | true     | false   | false         | false     | json:"phone,omitempty"             |          0 | phone of a user                |
| email             | string    | true   | true     | true     | false   | false         | false     | json:"email,omitempty"             |          0 | email of a user                |
| username          | string    | true   | true     | true     | false   | false         | false     | json:"username,omitempty"          |          2 | username of a user             |
| name              | string    | false  | false    | false    | true    | false         | false     | json:"name,omitempty"              |          0 | this field contains a name     |
|                   |           |        |          |          |         |               |           |                                    |            | that user set up               |
| bio               | string    | false  | false    | false    | true    | false         | false     | json:"bio,omitempty"               |          0 | this field a biography of a    |
|                   |           |        |          |          |         |               |           |                                    |            | user                           |
| avatar            | string    | false  | true     | true     | false   | false         | false     | json:"avatar,omitempty"            |          0 | a string contains link to a    |
|                   |           |        |          |          |         |               |           |                                    |            | profile pic                    |
| timezone          | string    | false  | false    | false    | true    | false         | false     | json:"timezone,omitempty"          |          0 | the timezone of a user         |
| is_active         | bool      | false  | false    | false    | true    | false         | false     | json:"is_active,omitempty"         |          0 | this field indicates that user |
|                   |           |        |          |          |         |               |           |                                    |            | finished his signup            |
| phone_verified    | bool      | false  | false    | false    | true    | false         | false     | json:"phone_verified,omitempty"    |          0 | this field indicates that      |
|                   |           |        |          |          |         |               |           |                                    |            | phone has been verified        |
| email_verified    | bool      | false  | false    | false    | true    | false         | false     | json:"email_verified,omitempty"    |          0 | this field indicates that      |
|                   |           |        |          |          |         |               |           |                                    |            | email has been verified        |
| last_login_at     | time.Time | false  | false    | false    | true    | false         | false     | json:"last_login_at,omitempty"     |          0 | the time when user was last    |
|                   |           |        |          |          |         |               |           |                                    |            | logged in                      |
| last_activity_at  | time.Time | false  | true     | true     | false   | false         | false     | json:"last_activity_at,omitempty"  |          0 | the time of user's last        |
|                   |           |        |          |          |         |               |           |                                    |            | activity                       |
| created_at        | time.Time | false  | false    | false    | true    | false         | false     | json:"created_at,omitempty"        |          0 | the time when user has been    |
|                   |           |        |          |          |         |               |           |                                    |            | created                        |
| updated_at        | time.Time | false  | false    | false    | true    | false         | false     | json:"updated_at,omitempty"        |          0 | the time when user was last    |
|                   |           |        |          |          |         |               |           |                                    |            | changed                        |
| bio_updated_at    | time.Time | false  | true     | true     | false   | false         | false     | json:"bio_updated_at,omitempty"    |          0 | the time when user's bio has   |
|                   |           |        |          |          |         |               |           |                                    |            | been changed                   |
| default_tenant_id | int64     | false  | true     | true     | false   | false         | false     | json:"default_tenant_id,omitempty" |          0 | default tenant id of user      |

## UserCredentials:

|      Field       |      Type      | Unique | Optional | Nillable | Default | UpdateDefault | Immutable |             StructTag             | Validators | Comment |
|---|---|---|---|---|---|---|---|---|---|---|
| id               | int            | false  | false    | false    | false   | false         | false     | json:"id,omitempty"               |          0 |         |
| created_at       | time.Time      | false  | false    | false    | true    | false         | true      | json:"created_at,omitempty"       |          0 |         |
| updated_at       | time.Time      | false  | false    | false    | true    | true          | false     | json:"updated_at,omitempty"       |          0 |         |
| deleted_at       | time.Time      | false  | true     | true     | false   | false         | false     | json:"deleted_at,omitempty"       |          0 |         |
| user_id          | int64          | false  | false    | false    | false   | false         | false     | json:"user_id,omitempty"          |          0 |         |
| external_user_id | int64          | false  | true     | true     | false   | false         | false     | json:"external_user_id,omitempty" |          0 |         |
| mail             | string         | false  | true     | true     | false   | false         | false     | json:"mail,omitempty"             |          0 |         |
| phone            | string         | false  | true     | true     | false   | false         | false     | json:"phone,omitempty"            |          0 |         |
| display_name     | string         | false  | true     | true     | false   | false         | false     | json:"display_name,omitempty"     |          0 |         |
| provider         | struc.Provider | false  | true     | true     | false   | false         | false     | json:"provider,omitempty"         |          0 |         |
| access_token     | string         | false  | false    | false    | false   | false         | false     | json:"access_token,omitempty"     |          0 |         |
| token_type       | string         | false  | true     | true     | false   | false         | false     | json:"token_type,omitempty"       |          0 |         |
| refresh_token    | string         | false  | true     | true     | false   | false         | false     | json:"refresh_token,omitempty"    |          0 |         |
| expires_at       | time.Time      | false  | true     | true     | false   | false         | false     | json:"expires_at,omitempty"       |          0 |         |


| Edge | Type | Inverse | BackRef | Relation | Unique | Optional | Comment |
|---|---|---|---|---|---|---|---|
| user | User | false   |         | M2O      | true   | false    |         |

## UserPrivacy:

|   Field    |         Type         | Unique | Optional | Nillable | Default | UpdateDefault | Immutable |          StructTag          | Validators | Comment |
|---|---|---|---|---|---|---|---|---|---|---|
| id         | int                  | false  | false    | false    | false   | false         | false     | json:"id,omitempty"         |          0 |         |
| user_id    | int64                | false  | false    | false    | false   | false         | false     | json:"user_id,omitempty"    |          0 |         |
| setting    | enum.PrivacySettings | false  | false    | false    | false   | false         | false     | json:"setting,omitempty"    |          0 |         |
| option     | enum.PrivacyOptions  | false  | false    | false    | false   | false         | false     | json:"option,omitempty"     |          0 |         |
| updated_at | time.Time            | false  | false    | false    | true    | false         | false     | json:"updated_at,omitempty" |          0 |         |


| Edge | Type | Inverse | BackRef | Relation | Unique | Optional | Comment |
|---|---|---|---|---|---|---|---|
| user | User | false   |         | M2O      | true   | false    |         |

## UserSettings:

|   Field    |     Type      | Unique | Optional | Nillable | Default | UpdateDefault | Immutable |          StructTag          | Validators | Comment |
|---|---|---|---|---|---|---|---|---|---|---|
| id         | int           | false  | false    | false    | false   | false         | false     | json:"id,omitempty"         |          0 |         |
| user_id    | int64         | false  | false    | false    | false   | false         | false     | json:"user_id,omitempty"    |          0 |         |
| setting    | enum.Settings | false  | false    | false    | false   | false         | false     | json:"setting,omitempty"    |          0 |         |
| value      | string        | false  | false    | false    | false   | false         | false     | json:"value,omitempty"      |          0 |         |
| updated_at | time.Time     | false  | false    | false    | true    | false         | false     | json:"updated_at,omitempty" |          0 |         |


| Edge | Type | Inverse | BackRef | Relation | Unique | Optional | Comment |
|---|---|---|---|---|---|---|---|
| user | User | false   |         | M2O      | true   | false    |         |

