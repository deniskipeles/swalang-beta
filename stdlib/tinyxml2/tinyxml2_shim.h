#ifndef TINYXML2_SHIM_H
#define TINYXML2_SHIM_H

#ifdef __cplusplus
extern "C" {
#endif

// Opaque pointers for the C++ types
typedef void* XMLDocumentPtr;
typedef void* XMLElementPtr;

// XMLDocument functions
XMLDocumentPtr XMLDocument_New();
void XMLDocument_Delete(XMLDocumentPtr doc);
int XMLDocument_Parse(XMLDocumentPtr doc, const char* xml);
XMLElementPtr XMLDocument_FirstChildElement(XMLDocumentPtr doc, const char* name);

// XMLElement functions
const char* XMLElement_Name(XMLElementPtr element);
const char* XMLElement_GetText(XMLElementPtr element);
const char* XMLElement_Attribute(XMLElementPtr element, const char* name);
XMLElementPtr XMLElement_FirstChildElement(XMLElementPtr element, const char* name);
XMLElementPtr XMLElement_NextSiblingElement(XMLElementPtr element, const char* name);

#ifdef __cplusplus
}
#endif

#endif // TINYXML2_SHIM_H