/**
 * @license Copyright (c) 2003-2016, CKSource - Frederico Knabben. All rights reserved.
 * For licensing, see LICENSE.md or http://ckeditor.com/license
 */

CKEDITOR.editorConfig = function( config ) {
	// Define changes to default configuration here. For example:
	// config.language = 'fr';
	// config.uiColor = '#AADC6E';
  config.toolbar = 'Full';
  config.toolbar_Full = [
    { name: 'document', items: ['CodeSnippet','Cut','-','Paste'] },
  ];
  //toolbar: ['source','save','/','save'],
  config.contentsCss = "/css/darkstrap.css";
  //extraPlugins: 'widget,codesnippet',
  config.imageUploadUrl = '/photo/';
  config.filebrowserImageBrowseUrl = '/photo/';
  config.filebrowserImageUploadUrl = '/photo/?';
};
