<!-- Generator: Widdershins v4.0.1 -->

<h1 id="api"> v0.0.1</h1>

> Scroll down for example requests and responses.

<h1 id="api-auth">Auth</h1>

## Auth_AuthByCode

<a id="opIdAuth_AuthByCode"></a>

`POST /v1/auth/code`

Auth by Code  
 after you authorized bu phone, you suppose to enter code that you've got here  
 request: id from AuthByPhone, code from message  
 returns: bearer token

> Body parameter

```json
{
  "userId": "string",
  "code": "string"
}
```

<h3 id="auth_authbycode-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[api.iam.v1.AuthByCodeRequest](#schemaapi.iam.v1.authbycoderequest)|true|none|

> Example responses

> 200 Response

```json
{
  "accessToken": "string",
  "refreshToken": "string"
}
```

<h3 id="auth_authbycode-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.TokenReply](#schemaapi.iam.v1.tokenreply)|

<aside class="success">
This operation does not require authentication
</aside>

## Auth_RefreshPersonalToken

<a id="opIdAuth_RefreshPersonalToken"></a>

`GET /v1/auth/personal`

> Example responses

> 200 Response

```json
{
  "accessToken": "string",
  "refreshToken": "string"
}
```

<h3 id="auth_refreshpersonaltoken-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.TokenReply](#schemaapi.iam.v1.tokenreply)|

<aside class="success">
This operation does not require authentication
</aside>

## Auth_AuthByPhone

<a id="opIdAuth_AuthByPhone"></a>

`POST /v1/auth/phone`

Auth by Phone  
 request: phone number  
 returns: id of newly created otp user

> Body parameter

```json
{
  "phone": "string"
}
```

<h3 id="auth_authbyphone-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[api.iam.v1.AuthByPhoneRequest](#schemaapi.iam.v1.authbyphonerequest)|true|none|

> Example responses

> 200 Response

```json
{
  "userId": "string"
}
```

<h3 id="auth_authbyphone-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.AuthByPhoneReply](#schemaapi.iam.v1.authbyphonereply)|

<aside class="success">
This operation does not require authentication
</aside>

## Auth_RefreshTenantToken

<a id="opIdAuth_RefreshTenantToken"></a>

`GET /v1/auth/tenant/{tenantId}`

<h3 id="auth_refreshtenanttoken-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|tenantId|path|string|true|none|

> Example responses

> 200 Response

```json
{
  "accessToken": "string",
  "refreshToken": "string"
}
```

<h3 id="auth_refreshtenanttoken-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.TokenReply](#schemaapi.iam.v1.tokenreply)|

<aside class="success">
This operation does not require authentication
</aside>

<h1 id="api-privacy">Privacy</h1>

## Privacy_GetPrivacy

<a id="opIdPrivacy_GetPrivacy"></a>

`GET /v1/users/me/privacy`

> Example responses

> 200 Response

```json
{
  "settings": {
    "property1": "string",
    "property2": "string"
  }
}
```

<h3 id="privacy_getprivacy-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.PrivacyReply](#schemaapi.iam.v1.privacyreply)|

<aside class="success">
This operation does not require authentication
</aside>

## Privacy_UpdatePrivacy

<a id="opIdPrivacy_UpdatePrivacy"></a>

`PUT /v1/users/me/privacy`

> Body parameter

```json
{
  "settings": {
    "property1": "string",
    "property2": "string"
  }
}
```

<h3 id="privacy_updateprivacy-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[api.iam.v1.PrivacyRequest](#schemaapi.iam.v1.privacyrequest)|true|none|

> Example responses

> 200 Response

```json
{
  "settings": {
    "property1": "string",
    "property2": "string"
  }
}
```

<h3 id="privacy_updateprivacy-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.PrivacyReply](#schemaapi.iam.v1.privacyreply)|

<aside class="success">
This operation does not require authentication
</aside>

<h1 id="api-settings">Settings</h1>

## Settings_GetSettings

<a id="opIdSettings_GetSettings"></a>

`GET /v1/users/me/settings`

> Example responses

> 200 Response

```json
{
  "settings": {
    "property1": "string",
    "property2": "string"
  }
}
```

<h3 id="settings_getsettings-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.SettingsReply](#schemaapi.iam.v1.settingsreply)|

<aside class="success">
This operation does not require authentication
</aside>

## Settings_UpdateSettings

<a id="opIdSettings_UpdateSettings"></a>

`PUT /v1/users/me/settings`

> Body parameter

```json
{
  "settings": {
    "property1": "string",
    "property2": "string"
  }
}
```

<h3 id="settings_updatesettings-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[api.iam.v1.SettingsRequest](#schemaapi.iam.v1.settingsrequest)|true|none|

> Example responses

> 200 Response

```json
{
  "settings": {
    "property1": "string",
    "property2": "string"
  }
}
```

<h3 id="settings_updatesettings-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.SettingsReply](#schemaapi.iam.v1.settingsreply)|

<aside class="success">
This operation does not require authentication
</aside>

<h1 id="api-users">Users</h1>

## Users_GetUsers

<a id="opIdUsers_GetUsers"></a>

`POST /v1/users/list`

search goes by all fileds  
 t.m if you declare labels,emails,ids this method will return all
 users that has these fields

> Body parameter

```json
{
  "ids": [
    "string"
  ],
  "phones": [
    "string"
  ],
  "emails": [
    "string"
  ]
}
```

<h3 id="users_getusers-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[api.iam.v1.GetUsersRequest](#schemaapi.iam.v1.getusersrequest)|true|none|

> Example responses

> 200 Response

```json
{
  "users": [
    {
      "id": "string",
      "phone": "string",
      "email": "string",
      "name": "string",
      "avatar": "string",
      "lastLoginAt": "string"
    }
  ]
}
```

<h3 id="users_getusers-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.GetUsersReply](#schemaapi.iam.v1.getusersreply)|

<aside class="success">
This operation does not require authentication
</aside>

## Users_GetOwnProfile

<a id="opIdUsers_GetOwnProfile"></a>

`GET /v1/users/me`

GetOwnProfile  
 This is self explanotory, returns own profile

> Example responses

> 200 Response

```json
{
  "user": {
    "id": "string",
    "name": "string",
    "bio": "string",
    "timezone": "string",
    "createdAt": "string",
    "updatedAt": "string",
    "lastLoginAt": "string",
    "isActive": true,
    "phone": "string",
    "email": "string",
    "avatar": "string",
    "bioUpdatedAt": "string",
    "contact": {
      "label": "string"
    }
  }
}
```

<h3 id="users_getownprofile-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.UserFullReply](#schemaapi.iam.v1.userfullreply)|

<aside class="success">
This operation does not require authentication
</aside>

## Users_UpdateOwnProfile

<a id="opIdUsers_UpdateOwnProfile"></a>

`POST /v1/users/me`

UpdateOwnProfile  
 This is self explanotory, update own profile

> Body parameter

```json
{
  "name": "string",
  "bio": "string",
  "avatar": "string",
  "timezone": "string"
}
```

<h3 id="users_updateownprofile-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[api.iam.v1.UpdateOwnProfileRequest](#schemaapi.iam.v1.updateownprofilerequest)|true|none|

> Example responses

> 200 Response

```json
{
  "user": {
    "id": "string",
    "name": "string",
    "bio": "string",
    "timezone": "string",
    "createdAt": "string",
    "updatedAt": "string",
    "lastLoginAt": "string",
    "isActive": true,
    "phone": "string",
    "email": "string",
    "avatar": "string",
    "bioUpdatedAt": "string",
    "contact": {
      "label": "string"
    }
  }
}
```

<h3 id="users_updateownprofile-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.UserFullReply](#schemaapi.iam.v1.userfullreply)|

<aside class="success">
This operation does not require authentication
</aside>

## Users_DeleteOwnProfile

<a id="opIdUsers_DeleteOwnProfile"></a>

`DELETE /v1/users/me`

DeleteOwnProfile  
 This is self explanotory, delete own profile

> Example responses

> 200 Response

```json
{}
```

<h3 id="users_deleteownprofile-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.EmptyReply](#schemaapi.iam.v1.emptyreply)|

<aside class="success">
This operation does not require authentication
</aside>

## Users_GetUserByFilter

<a id="opIdUsers_GetUserByFilter"></a>

`POST /v1/users/user`

in case of search by phone, search.email should not be present  
 in case of search by email, search.phone should not be present

> Body parameter

```json
{
  "search": {
    "phone": "string",
    "email": "string"
  }
}
```

<h3 id="users_getuserbyfilter-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[api.iam.v1.GetUserByFilterRequest](#schemaapi.iam.v1.getuserbyfilterrequest)|true|none|

> Example responses

> 200 Response

```json
{
  "user": {
    "id": "string",
    "phone": "string",
    "email": "string",
    "name": "string",
    "avatar": "string",
    "lastLoginAt": "string"
  }
}
```

<h3 id="users_getuserbyfilter-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.UserReply](#schemaapi.iam.v1.userreply)|

<aside class="success">
This operation does not require authentication
</aside>

## Users_GetUserByFilterFull

<a id="opIdUsers_GetUserByFilterFull"></a>

`POST /v1/users/user/full`

in case of search by phone, search.email should not be present  
 in case of search by email, search.phone should not be present

> Body parameter

```json
{
  "search": {
    "phone": "string",
    "email": "string"
  }
}
```

<h3 id="users_getuserbyfilterfull-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[api.iam.v1.GetUserByFilterRequest](#schemaapi.iam.v1.getuserbyfilterrequest)|true|none|

> Example responses

> 200 Response

```json
{
  "user": {
    "id": "string",
    "name": "string",
    "bio": "string",
    "timezone": "string",
    "createdAt": "string",
    "updatedAt": "string",
    "lastLoginAt": "string",
    "isActive": true,
    "phone": "string",
    "email": "string",
    "avatar": "string",
    "bioUpdatedAt": "string",
    "contact": {
      "label": "string"
    }
  }
}
```

<h3 id="users_getuserbyfilterfull-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.UserFullReply](#schemaapi.iam.v1.userfullreply)|

<aside class="success">
This operation does not require authentication
</aside>

## Users_GetUser

<a id="opIdUsers_GetUser"></a>

`GET /v1/users/{userId}`

GetUser  
 get single user

<h3 id="users_getuser-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|userId|path|string|true|none|

> Example responses

> 200 Response

```json
{
  "user": {
    "id": "string",
    "phone": "string",
    "email": "string",
    "name": "string",
    "avatar": "string",
    "lastLoginAt": "string"
  }
}
```

<h3 id="users_getuser-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.UserReply](#schemaapi.iam.v1.userreply)|

<aside class="success">
This operation does not require authentication
</aside>

## Users_GetUserFull

<a id="opIdUsers_GetUserFull"></a>

`GET /v1/users/{userId}/full`

GetUserFull  
 Returns full information about user  
 Request: userId of seeking user

<h3 id="users_getuserfull-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|userId|path|string|true|none|

> Example responses

> 200 Response

```json
{
  "user": {
    "id": "string",
    "name": "string",
    "bio": "string",
    "timezone": "string",
    "createdAt": "string",
    "updatedAt": "string",
    "lastLoginAt": "string",
    "isActive": true,
    "phone": "string",
    "email": "string",
    "avatar": "string",
    "bioUpdatedAt": "string",
    "contact": {
      "label": "string"
    }
  }
}
```

<h3 id="users_getuserfull-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[api.iam.v1.UserFullReply](#schemaapi.iam.v1.userfullreply)|

<aside class="success">
This operation does not require authentication
</aside>

# Schemas

<h2 id="tocS_api.iam.v1.AuthByCodeRequest">api.iam.v1.AuthByCodeRequest</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.authbycoderequest"></a>
<a id="schema_api.iam.v1.AuthByCodeRequest"></a>
<a id="tocSapi.iam.v1.authbycoderequest"></a>
<a id="tocsapi.iam.v1.authbycoderequest"></a>

```json
{
  "userId": "string",
  "code": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|userId|string|false|none|none|
|code|string|false|none|none|

<h2 id="tocS_api.iam.v1.AuthByPhoneReply">api.iam.v1.AuthByPhoneReply</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.authbyphonereply"></a>
<a id="schema_api.iam.v1.AuthByPhoneReply"></a>
<a id="tocSapi.iam.v1.authbyphonereply"></a>
<a id="tocsapi.iam.v1.authbyphonereply"></a>

```json
{
  "userId": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|userId|string|false|none|none|

<h2 id="tocS_api.iam.v1.AuthByPhoneRequest">api.iam.v1.AuthByPhoneRequest</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.authbyphonerequest"></a>
<a id="schema_api.iam.v1.AuthByPhoneRequest"></a>
<a id="tocSapi.iam.v1.authbyphonerequest"></a>
<a id="tocsapi.iam.v1.authbyphonerequest"></a>

```json
{
  "phone": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|phone|string|false|none|none|

<h2 id="tocS_api.iam.v1.Contact">api.iam.v1.Contact</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.contact"></a>
<a id="schema_api.iam.v1.Contact"></a>
<a id="tocSapi.iam.v1.contact"></a>
<a id="tocsapi.iam.v1.contact"></a>

```json
{
  "label": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|label|string|false|none|none|

<h2 id="tocS_api.iam.v1.EmptyReply">api.iam.v1.EmptyReply</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.emptyreply"></a>
<a id="schema_api.iam.v1.EmptyReply"></a>
<a id="tocSapi.iam.v1.emptyreply"></a>
<a id="tocsapi.iam.v1.emptyreply"></a>

```json
{}

```

### Properties

*None*

<h2 id="tocS_api.iam.v1.GetUserByFilterRequest">api.iam.v1.GetUserByFilterRequest</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.getuserbyfilterrequest"></a>
<a id="schema_api.iam.v1.GetUserByFilterRequest"></a>
<a id="tocSapi.iam.v1.getuserbyfilterrequest"></a>
<a id="tocsapi.iam.v1.getuserbyfilterrequest"></a>

```json
{
  "search": {
    "phone": "string",
    "email": "string"
  }
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|search|[api.iam.v1.SearchFilter](#schemaapi.iam.v1.searchfilter)|false|none|none|

<h2 id="tocS_api.iam.v1.GetUsersReply">api.iam.v1.GetUsersReply</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.getusersreply"></a>
<a id="schema_api.iam.v1.GetUsersReply"></a>
<a id="tocSapi.iam.v1.getusersreply"></a>
<a id="tocsapi.iam.v1.getusersreply"></a>

```json
{
  "users": [
    {
      "id": "string",
      "phone": "string",
      "email": "string",
      "name": "string",
      "avatar": "string",
      "lastLoginAt": "string"
    }
  ]
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|users|[[api.iam.v1.UserShort](#schemaapi.iam.v1.usershort)]|false|none|none|

<h2 id="tocS_api.iam.v1.GetUsersRequest">api.iam.v1.GetUsersRequest</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.getusersrequest"></a>
<a id="schema_api.iam.v1.GetUsersRequest"></a>
<a id="tocSapi.iam.v1.getusersrequest"></a>
<a id="tocsapi.iam.v1.getusersrequest"></a>

```json
{
  "ids": [
    "string"
  ],
  "phones": [
    "string"
  ],
  "emails": [
    "string"
  ]
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|ids|[string]|false|none|none|
|phones|[string]|false|none|none|
|emails|[string]|false|none|none|

<h2 id="tocS_api.iam.v1.PrivacyReply">api.iam.v1.PrivacyReply</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.privacyreply"></a>
<a id="schema_api.iam.v1.PrivacyReply"></a>
<a id="tocSapi.iam.v1.privacyreply"></a>
<a id="tocsapi.iam.v1.privacyreply"></a>

```json
{
  "settings": {
    "property1": "string",
    "property2": "string"
  }
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|settings|object|false|none|none|
|» **additionalProperties**|string|false|none|none|

<h2 id="tocS_api.iam.v1.PrivacyRequest">api.iam.v1.PrivacyRequest</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.privacyrequest"></a>
<a id="schema_api.iam.v1.PrivacyRequest"></a>
<a id="tocSapi.iam.v1.privacyrequest"></a>
<a id="tocsapi.iam.v1.privacyrequest"></a>

```json
{
  "settings": {
    "property1": "string",
    "property2": "string"
  }
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|settings|object|false|none|none|
|» **additionalProperties**|string|false|none|none|

<h2 id="tocS_api.iam.v1.SearchFilter">api.iam.v1.SearchFilter</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.searchfilter"></a>
<a id="schema_api.iam.v1.SearchFilter"></a>
<a id="tocSapi.iam.v1.searchfilter"></a>
<a id="tocsapi.iam.v1.searchfilter"></a>

```json
{
  "phone": "string",
  "email": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|phone|string|false|none|none|
|email|string|false|none|none|

<h2 id="tocS_api.iam.v1.SettingsReply">api.iam.v1.SettingsReply</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.settingsreply"></a>
<a id="schema_api.iam.v1.SettingsReply"></a>
<a id="tocSapi.iam.v1.settingsreply"></a>
<a id="tocsapi.iam.v1.settingsreply"></a>

```json
{
  "settings": {
    "property1": "string",
    "property2": "string"
  }
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|settings|object|false|none|none|
|» **additionalProperties**|string|false|none|none|

<h2 id="tocS_api.iam.v1.SettingsRequest">api.iam.v1.SettingsRequest</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.settingsrequest"></a>
<a id="schema_api.iam.v1.SettingsRequest"></a>
<a id="tocSapi.iam.v1.settingsrequest"></a>
<a id="tocsapi.iam.v1.settingsrequest"></a>

```json
{
  "settings": {
    "property1": "string",
    "property2": "string"
  }
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|settings|object|false|none|none|
|» **additionalProperties**|string|false|none|none|

<h2 id="tocS_api.iam.v1.TokenReply">api.iam.v1.TokenReply</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.tokenreply"></a>
<a id="schema_api.iam.v1.TokenReply"></a>
<a id="tocSapi.iam.v1.tokenreply"></a>
<a id="tocsapi.iam.v1.tokenreply"></a>

```json
{
  "accessToken": "string",
  "refreshToken": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|accessToken|string|false|none|none|
|refreshToken|string|false|none|none|

<h2 id="tocS_api.iam.v1.UpdateOwnProfileRequest">api.iam.v1.UpdateOwnProfileRequest</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.updateownprofilerequest"></a>
<a id="schema_api.iam.v1.UpdateOwnProfileRequest"></a>
<a id="tocSapi.iam.v1.updateownprofilerequest"></a>
<a id="tocsapi.iam.v1.updateownprofilerequest"></a>

```json
{
  "name": "string",
  "bio": "string",
  "avatar": "string",
  "timezone": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|name|string|false|none|none|
|bio|string|false|none|none|
|avatar|string|false|none|none|
|timezone|string|false|none|none|

<h2 id="tocS_api.iam.v1.User">api.iam.v1.User</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.user"></a>
<a id="schema_api.iam.v1.User"></a>
<a id="tocSapi.iam.v1.user"></a>
<a id="tocsapi.iam.v1.user"></a>

```json
{
  "id": "string",
  "name": "string",
  "bio": "string",
  "timezone": "string",
  "createdAt": "string",
  "updatedAt": "string",
  "lastLoginAt": "string",
  "isActive": true,
  "phone": "string",
  "email": "string",
  "avatar": "string",
  "bioUpdatedAt": "string",
  "contact": {
    "label": "string"
  }
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|id|string|false|none|none|
|name|string|false|none|none|
|bio|string|false|none|none|
|timezone|string|false|none|none|
|createdAt|string|false|none|none|
|updatedAt|string|false|none|none|
|lastLoginAt|string|false|none|none|
|isActive|boolean|false|none|none|
|phone|string|false|none|none|
|email|string|false|none|none|
|avatar|string|false|none|none|
|bioUpdatedAt|string|false|none|none|
|contact|[api.iam.v1.Contact](#schemaapi.iam.v1.contact)|false|none|field containing tied contact info|

<h2 id="tocS_api.iam.v1.UserFullReply">api.iam.v1.UserFullReply</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.userfullreply"></a>
<a id="schema_api.iam.v1.UserFullReply"></a>
<a id="tocSapi.iam.v1.userfullreply"></a>
<a id="tocsapi.iam.v1.userfullreply"></a>

```json
{
  "user": {
    "id": "string",
    "name": "string",
    "bio": "string",
    "timezone": "string",
    "createdAt": "string",
    "updatedAt": "string",
    "lastLoginAt": "string",
    "isActive": true,
    "phone": "string",
    "email": "string",
    "avatar": "string",
    "bioUpdatedAt": "string",
    "contact": {
      "label": "string"
    }
  }
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|user|[api.iam.v1.User](#schemaapi.iam.v1.user)|false|none|none|

<h2 id="tocS_api.iam.v1.UserReply">api.iam.v1.UserReply</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.userreply"></a>
<a id="schema_api.iam.v1.UserReply"></a>
<a id="tocSapi.iam.v1.userreply"></a>
<a id="tocsapi.iam.v1.userreply"></a>

```json
{
  "user": {
    "id": "string",
    "phone": "string",
    "email": "string",
    "name": "string",
    "avatar": "string",
    "lastLoginAt": "string"
  }
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|user|[api.iam.v1.UserShort](#schemaapi.iam.v1.usershort)|false|none|none|

<h2 id="tocS_api.iam.v1.UserShort">api.iam.v1.UserShort</h2>
<!-- backwards compatibility -->
<a id="schemaapi.iam.v1.usershort"></a>
<a id="schema_api.iam.v1.UserShort"></a>
<a id="tocSapi.iam.v1.usershort"></a>
<a id="tocsapi.iam.v1.usershort"></a>

```json
{
  "id": "string",
  "phone": "string",
  "email": "string",
  "name": "string",
  "avatar": "string",
  "lastLoginAt": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|id|string|false|none|none|
|phone|string|false|none|none|
|email|string|false|none|none|
|name|string|false|none|none|
|avatar|string|false|none|none|
|lastLoginAt|string|false|none|none|

