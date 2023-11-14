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
|body|body|[iam.v1.AuthByCodeRequest](#schemaiam.v1.authbycoderequest)|true|none|

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
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.TokenReply](#schemaiam.v1.tokenreply)|

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
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.TokenReply](#schemaiam.v1.tokenreply)|

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
|body|body|[iam.v1.AuthByPhoneRequest](#schemaiam.v1.authbyphonerequest)|true|none|

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
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.AuthByPhoneReply](#schemaiam.v1.authbyphonereply)|

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
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.TokenReply](#schemaiam.v1.tokenreply)|

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
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.PrivacyReply](#schemaiam.v1.privacyreply)|

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
|body|body|[iam.v1.PrivacyRequest](#schemaiam.v1.privacyrequest)|true|none|

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
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.PrivacyReply](#schemaiam.v1.privacyreply)|

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
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.SettingsReply](#schemaiam.v1.settingsreply)|

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
|body|body|[iam.v1.SettingsRequest](#schemaiam.v1.settingsrequest)|true|none|

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
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.SettingsReply](#schemaiam.v1.settingsreply)|

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
  ],
  "withRelation": true
}
```

<h3 id="users_getusers-parameters">Parameters</h3>

|Name|In|Type|Required|Description|
|---|---|---|---|---|
|body|body|[iam.v1.GetUsersRequest](#schemaiam.v1.getusersrequest)|true|none|

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
      "lastLoginAt": "string",
      "relation": {
        "isBlocked": true,
        "isMuted": true
      }
    }
  ]
}
```

<h3 id="users_getusers-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.GetUsersReply](#schemaiam.v1.getusersreply)|

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
    },
    "relation": {
      "isBlocked": true,
      "isMuted": true
    },
    "directChat": {
      "chatId": "string",
      "status": "string",
      "role": "string",
      "isPinned": true,
      "isMuted": true,
      "mutedTill": "string",
      "archivedAt": "string",
      "autoSave": true
    }
  }
}
```

<h3 id="users_getownprofile-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.UserFullReply](#schemaiam.v1.userfullreply)|

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
|body|body|[iam.v1.UpdateOwnProfileRequest](#schemaiam.v1.updateownprofilerequest)|true|none|

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
    },
    "relation": {
      "isBlocked": true,
      "isMuted": true
    },
    "directChat": {
      "chatId": "string",
      "status": "string",
      "role": "string",
      "isPinned": true,
      "isMuted": true,
      "mutedTill": "string",
      "archivedAt": "string",
      "autoSave": true
    }
  }
}
```

<h3 id="users_updateownprofile-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.UserFullReply](#schemaiam.v1.userfullreply)|

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
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.EmptyReply](#schemaiam.v1.emptyreply)|

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
|body|body|[iam.v1.GetUserByFilterRequest](#schemaiam.v1.getuserbyfilterrequest)|true|none|

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
    "lastLoginAt": "string",
    "relation": {
      "isBlocked": true,
      "isMuted": true
    }
  }
}
```

<h3 id="users_getuserbyfilter-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.UserReply](#schemaiam.v1.userreply)|

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
|body|body|[iam.v1.GetUserByFilterRequest](#schemaiam.v1.getuserbyfilterrequest)|true|none|

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
    },
    "relation": {
      "isBlocked": true,
      "isMuted": true
    },
    "directChat": {
      "chatId": "string",
      "status": "string",
      "role": "string",
      "isPinned": true,
      "isMuted": true,
      "mutedTill": "string",
      "archivedAt": "string",
      "autoSave": true
    }
  }
}
```

<h3 id="users_getuserbyfilterfull-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.UserFullReply](#schemaiam.v1.userfullreply)|

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
    "lastLoginAt": "string",
    "relation": {
      "isBlocked": true,
      "isMuted": true
    }
  }
}
```

<h3 id="users_getuser-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.UserReply](#schemaiam.v1.userreply)|

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
    },
    "relation": {
      "isBlocked": true,
      "isMuted": true
    },
    "directChat": {
      "chatId": "string",
      "status": "string",
      "role": "string",
      "isPinned": true,
      "isMuted": true,
      "mutedTill": "string",
      "archivedAt": "string",
      "autoSave": true
    }
  }
}
```

<h3 id="users_getuserfull-responses">Responses</h3>

|Status|Meaning|Description|Schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|[iam.v1.UserFullReply](#schemaiam.v1.userfullreply)|

<aside class="success">
This operation does not require authentication
</aside>

# Schemas

<h2 id="tocS_iam.v1.AuthByCodeRequest">iam.v1.AuthByCodeRequest</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.authbycoderequest"></a>
<a id="schema_iam.v1.AuthByCodeRequest"></a>
<a id="tocSiam.v1.authbycoderequest"></a>
<a id="tocsiam.v1.authbycoderequest"></a>

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

<h2 id="tocS_iam.v1.AuthByPhoneReply">iam.v1.AuthByPhoneReply</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.authbyphonereply"></a>
<a id="schema_iam.v1.AuthByPhoneReply"></a>
<a id="tocSiam.v1.authbyphonereply"></a>
<a id="tocsiam.v1.authbyphonereply"></a>

```json
{
  "userId": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|userId|string|false|none|none|

<h2 id="tocS_iam.v1.AuthByPhoneRequest">iam.v1.AuthByPhoneRequest</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.authbyphonerequest"></a>
<a id="schema_iam.v1.AuthByPhoneRequest"></a>
<a id="tocSiam.v1.authbyphonerequest"></a>
<a id="tocsiam.v1.authbyphonerequest"></a>

```json
{
  "phone": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|phone|string|false|none|none|

<h2 id="tocS_iam.v1.Contact">iam.v1.Contact</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.contact"></a>
<a id="schema_iam.v1.Contact"></a>
<a id="tocSiam.v1.contact"></a>
<a id="tocsiam.v1.contact"></a>

```json
{
  "label": "string"
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|label|string|false|none|none|

<h2 id="tocS_iam.v1.DirectChat">iam.v1.DirectChat</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.directchat"></a>
<a id="schema_iam.v1.DirectChat"></a>
<a id="tocSiam.v1.directchat"></a>
<a id="tocsiam.v1.directchat"></a>

```json
{
  "chatId": "string",
  "status": "string",
  "role": "string",
  "isPinned": true,
  "isMuted": true,
  "mutedTill": "string",
  "archivedAt": "string",
  "autoSave": true
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|chatId|string|false|none|none|
|status|string|false|none|none|
|role|string|false|none|none|
|isPinned|boolean|false|none|none|
|isMuted|boolean|false|none|none|
|mutedTill|string|false|none|none|
|archivedAt|string|false|none|none|
|autoSave|boolean|false|none|none|

<h2 id="tocS_iam.v1.EmptyReply">iam.v1.EmptyReply</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.emptyreply"></a>
<a id="schema_iam.v1.EmptyReply"></a>
<a id="tocSiam.v1.emptyreply"></a>
<a id="tocsiam.v1.emptyreply"></a>

```json
{}

```

### Properties

*None*

<h2 id="tocS_iam.v1.GetUserByFilterRequest">iam.v1.GetUserByFilterRequest</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.getuserbyfilterrequest"></a>
<a id="schema_iam.v1.GetUserByFilterRequest"></a>
<a id="tocSiam.v1.getuserbyfilterrequest"></a>
<a id="tocsiam.v1.getuserbyfilterrequest"></a>

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
|search|[iam.v1.SearchFilter](#schemaiam.v1.searchfilter)|false|none|none|

<h2 id="tocS_iam.v1.GetUsersReply">iam.v1.GetUsersReply</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.getusersreply"></a>
<a id="schema_iam.v1.GetUsersReply"></a>
<a id="tocSiam.v1.getusersreply"></a>
<a id="tocsiam.v1.getusersreply"></a>

```json
{
  "users": [
    {
      "id": "string",
      "phone": "string",
      "email": "string",
      "name": "string",
      "avatar": "string",
      "lastLoginAt": "string",
      "relation": {
        "isBlocked": true,
        "isMuted": true
      }
    }
  ]
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|users|[[iam.v1.UserShort](#schemaiam.v1.usershort)]|false|none|none|

<h2 id="tocS_iam.v1.GetUsersRequest">iam.v1.GetUsersRequest</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.getusersrequest"></a>
<a id="schema_iam.v1.GetUsersRequest"></a>
<a id="tocSiam.v1.getusersrequest"></a>
<a id="tocsiam.v1.getusersrequest"></a>

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
  ],
  "withRelation": true
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|ids|[string]|false|none|none|
|phones|[string]|false|none|none|
|emails|[string]|false|none|none|
|withRelation|boolean|false|none|none|

<h2 id="tocS_iam.v1.PrivacyReply">iam.v1.PrivacyReply</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.privacyreply"></a>
<a id="schema_iam.v1.PrivacyReply"></a>
<a id="tocSiam.v1.privacyreply"></a>
<a id="tocsiam.v1.privacyreply"></a>

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

<h2 id="tocS_iam.v1.PrivacyRequest">iam.v1.PrivacyRequest</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.privacyrequest"></a>
<a id="schema_iam.v1.PrivacyRequest"></a>
<a id="tocSiam.v1.privacyrequest"></a>
<a id="tocsiam.v1.privacyrequest"></a>

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

<h2 id="tocS_iam.v1.Relation">iam.v1.Relation</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.relation"></a>
<a id="schema_iam.v1.Relation"></a>
<a id="tocSiam.v1.relation"></a>
<a id="tocsiam.v1.relation"></a>

```json
{
  "isBlocked": true,
  "isMuted": true
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|isBlocked|boolean|false|none|none|
|isMuted|boolean|false|none|none|

<h2 id="tocS_iam.v1.SearchFilter">iam.v1.SearchFilter</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.searchfilter"></a>
<a id="schema_iam.v1.SearchFilter"></a>
<a id="tocSiam.v1.searchfilter"></a>
<a id="tocsiam.v1.searchfilter"></a>

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

<h2 id="tocS_iam.v1.SettingsReply">iam.v1.SettingsReply</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.settingsreply"></a>
<a id="schema_iam.v1.SettingsReply"></a>
<a id="tocSiam.v1.settingsreply"></a>
<a id="tocsiam.v1.settingsreply"></a>

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

<h2 id="tocS_iam.v1.SettingsRequest">iam.v1.SettingsRequest</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.settingsrequest"></a>
<a id="schema_iam.v1.SettingsRequest"></a>
<a id="tocSiam.v1.settingsrequest"></a>
<a id="tocsiam.v1.settingsrequest"></a>

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

<h2 id="tocS_iam.v1.TokenReply">iam.v1.TokenReply</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.tokenreply"></a>
<a id="schema_iam.v1.TokenReply"></a>
<a id="tocSiam.v1.tokenreply"></a>
<a id="tocsiam.v1.tokenreply"></a>

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

<h2 id="tocS_iam.v1.UpdateOwnProfileRequest">iam.v1.UpdateOwnProfileRequest</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.updateownprofilerequest"></a>
<a id="schema_iam.v1.UpdateOwnProfileRequest"></a>
<a id="tocSiam.v1.updateownprofilerequest"></a>
<a id="tocsiam.v1.updateownprofilerequest"></a>

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

<h2 id="tocS_iam.v1.User">iam.v1.User</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.user"></a>
<a id="schema_iam.v1.User"></a>
<a id="tocSiam.v1.user"></a>
<a id="tocsiam.v1.user"></a>

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
  },
  "relation": {
    "isBlocked": true,
    "isMuted": true
  },
  "directChat": {
    "chatId": "string",
    "status": "string",
    "role": "string",
    "isPinned": true,
    "isMuted": true,
    "mutedTill": "string",
    "archivedAt": "string",
    "autoSave": true
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
|contact|[iam.v1.Contact](#schemaiam.v1.contact)|false|none|field contains contact info|
|relation|[iam.v1.Relation](#schemaiam.v1.relation)|false|none|field contains relation info|
|directChat|[iam.v1.DirectChat](#schemaiam.v1.directchat)|false|none|field contains directChat info|

<h2 id="tocS_iam.v1.UserFullReply">iam.v1.UserFullReply</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.userfullreply"></a>
<a id="schema_iam.v1.UserFullReply"></a>
<a id="tocSiam.v1.userfullreply"></a>
<a id="tocsiam.v1.userfullreply"></a>

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
    },
    "relation": {
      "isBlocked": true,
      "isMuted": true
    },
    "directChat": {
      "chatId": "string",
      "status": "string",
      "role": "string",
      "isPinned": true,
      "isMuted": true,
      "mutedTill": "string",
      "archivedAt": "string",
      "autoSave": true
    }
  }
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|user|[iam.v1.User](#schemaiam.v1.user)|false|none|none|

<h2 id="tocS_iam.v1.UserReply">iam.v1.UserReply</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.userreply"></a>
<a id="schema_iam.v1.UserReply"></a>
<a id="tocSiam.v1.userreply"></a>
<a id="tocsiam.v1.userreply"></a>

```json
{
  "user": {
    "id": "string",
    "phone": "string",
    "email": "string",
    "name": "string",
    "avatar": "string",
    "lastLoginAt": "string",
    "relation": {
      "isBlocked": true,
      "isMuted": true
    }
  }
}

```

### Properties

|Name|Type|Required|Restrictions|Description|
|---|---|---|---|---|
|user|[iam.v1.UserShort](#schemaiam.v1.usershort)|false|none|none|

<h2 id="tocS_iam.v1.UserShort">iam.v1.UserShort</h2>
<!-- backwards compatibility -->
<a id="schemaiam.v1.usershort"></a>
<a id="schema_iam.v1.UserShort"></a>
<a id="tocSiam.v1.usershort"></a>
<a id="tocsiam.v1.usershort"></a>

```json
{
  "id": "string",
  "phone": "string",
  "email": "string",
  "name": "string",
  "avatar": "string",
  "lastLoginAt": "string",
  "relation": {
    "isBlocked": true,
    "isMuted": true
  }
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
|relation|[iam.v1.Relation](#schemaiam.v1.relation)|false|none|none|

