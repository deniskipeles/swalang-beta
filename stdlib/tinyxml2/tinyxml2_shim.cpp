#include "tinyxml2.h"
#include "tinyxml2_shim.h"

using namespace tinyxml2;

extern "C" {

XMLDocumentPtr XMLDocument_New() {
    return new XMLDocument();
}

void XMLDocument_Delete(XMLDocumentPtr doc) {
    delete static_cast<XMLDocument*>(doc);
}

int XMLDocument_Parse(XMLDocumentPtr doc, const char* xml) {
    return static_cast<XMLDocument*>(doc)->Parse(xml);
}

XMLElementPtr XMLDocument_FirstChildElement(XMLDocumentPtr doc, const char* name) {
    return static_cast<XMLDocument*>(doc)->FirstChildElement(name);
}

const char* XMLElement_Name(XMLElementPtr element) {
    return static_cast<XMLElement*>(element)->Name();
}

const char* XMLElement_GetText(XMLElementPtr element) {
    return static_cast<XMLElement*>(element)->GetText();
}

const char* XMLElement_Attribute(XMLElementPtr element, const char* name) {
    return static_cast<XMLElement*>(element)->Attribute(name);
}

XMLElementPtr XMLElement_FirstChildElement(XMLElementPtr element, const char* name) {
    return static_cast<XMLElement*>(element)->FirstChildElement(name);
}

XMLElementPtr XMLElement_NextSiblingElement(XMLElementPtr element, const char* name) {
    return static_cast<XMLElement*>(element)->NextSiblingElement(name);
}

} // extern "C"