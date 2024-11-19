$(document).ready(function () {
	// 初始化设置对象
	var settings = JSON.parse(localStorage.getItem('settings')) || {};
	if (!settings.OneHub) {
		settings.OneHub = {
			modelEndpoint: '',
			apiKey: '',
			oneHubToken: '',
			models: []
		};
		localStorage.setItem('settings', JSON.stringify(settings));
	}
	
	var selectedModelInfo = JSON.parse(localStorage.getItem('selectedModelInfo')) || {
		modelId: "gpt-3.5-turbo",
		provider: "OneHub",
		modelEndpoint: '',
		apiKey: '',
		oneHubToken: ''
	};

	function initSelect2() {
		$('#model').select2({
			placeholder: '选择模型',
			minimumResultsForSearch: Infinity,
			width: '100%',
			multiple: true,
		});
		
		// 创建一个包装容器
		var wrapper = $('<div>', {
			class: 'model-select-wrapper',
			style: 'display: flex; align-items: center; width: 60%;'
		});
		
		// 将select2移动到包装容器中
		$('#model').next('.select2').detach().appendTo(wrapper);
		
		// 创建全选按钮并添加到包装容器
		var selectAllBtn = $('<button>', {
			class: 'btn btn-default select-all-btn',
			text: '全选',
			style: 'margin-left: 8px; padding: 5px 10px; min-width: 50px;'
		});
		
		wrapper.append(selectAllBtn);
		
		// 将包装容器插入到原select元素后面
		$('#model').after(wrapper);
		
		// 全选按钮点击事件
		selectAllBtn.click(function() {
			var allOptions = $('#model option').map(function() {
				return $(this).val();
			}).get();
			
			if($('#model').val().length === allOptions.length) {
				// 如果已经全选，则清空选择
				$('#model').val(null).trigger('change');
			} else {
				// 否则全选
				$('#model').val(allOptions).trigger('change');
			}
			
			// 触发保存设置
			var provider = $('#serviceProvider').val();
			saveProviderSettings(provider);
			updateSingleModelOptions();

			// 调用本地代理接口更新模型选择
			updateModelSelect(settings[provider].models || []);
		});
	}

	function updateProviderSettings(provider) {
		var providerSettings = settings[provider] || {};
		$('#modelEndpoint').val(providerSettings.modelEndpoint || '');
		$('#apiKey').val(providerSettings.apiKey || '');
		$('#oneHubToken').val(providerSettings.oneHubToken || '');
		updateModelSelect(providerSettings.models || []);
	}

	function updateModelSelect(selectedModels) {
		var modelEndpoint = $('#modelEndpoint').val();
		var oneHubToken = $('#oneHubToken').val();
		
		if (!modelEndpoint) {
			// 如果没有设置modelEndpoint，使用空列表
			$('#model').empty().trigger('change');
			return;
		}

		// 使用本地代理接口
		var apiUrl = '/api/proxy/models';
		
		// 发送GET请求获取模型列表
		$.ajax({
			url: apiUrl,
			method: 'GET',
			data: {
				endpoint: modelEndpoint,
				oneHubToken: oneHubToken
			},
			success: function(response) {
				var modelOptions = response.data.map(function(model) {
					return new Option(
						model.id + (model.owned_by ? ' (' + model.owned_by + ')' : ''),
						model.id,
						selectedModels.includes(model.id),
						selectedModels.includes(model.id)
					);
				});

				$('#model').empty().append(modelOptions).trigger('change');
			},
			error: function(xhr, status, error) {
				console.error('Error fetching models:', error);
				// 发生错误时清空选项
				$('#model').empty().trigger('change');
			}
		});
	}

	function saveProviderSettings(provider) {
		// 确保获取所有输入值
		var currentSettings = settings[provider] || {};
		settings[provider] = {
			...currentSettings,
			modelEndpoint: $('#modelEndpoint').val() || '',
			apiKey: $('#apiKey').val() || '',
			oneHubToken: $('#oneHubToken').val() || '',
			models: $('#model').val() || []
		};
		
		// 保存到localStorage
		localStorage.setItem('settings', JSON.stringify(settings));
		console.log('Saved settings:', settings); // 添加日志以便调试
		
		// 同时更新当前选中模型的信息
		var currentModelValue = $('#singleModel').val();
		if (currentModelValue) {
			storeSelectedModelInfo(currentModelValue);
		}
	}

	function updateSingleModelOptions() {
		var $singleModelSelect = $('#singleModel');
		$singleModelSelect.empty();

		$singleModelSelect.append(
			$('<option>', {
				value: '{"modelId":"gpt-3.5-turbo","provider":"OneHub"}',
				text: '请选择模型'
			})
		);

		Object.keys(settings).forEach(function (provider) {
			(settings[provider].models || []).forEach(function (modelId) {
				var modelValue = JSON.stringify({
					modelId: modelId,
					provider: provider,
				});
				$singleModelSelect.append(
					// new Option(modelId + '🏷️' + provider, modelValue)
					new Option('🏷️' + modelId, modelValue)
				);
			});
		});

		setSelectedModel();
	}

	function setSelectedModel() {
		var $singleModelSelect = $('#singleModel');
		var isValid =
			settings[selectedModelInfo.provider] &&
			settings[selectedModelInfo.provider].models.includes(
				selectedModelInfo.modelId
			);

		// 使用JSON.stringify确保option的value和selectedModelInfo匹配
		var selectedModelValue = JSON.stringify({
			modelId: selectedModelInfo.modelId,
			provider: selectedModelInfo.provider,
		});

		if (isValid || selectedModelInfo.provider === 'fucker') {
			$singleModelSelect.val(selectedModelValue).trigger('change');
		} else {
			$singleModelSelect
				.val('{"modelId":"gpt-3.5-turbo","provider":"OneHub"}')
				.trigger('change');
			storeSelectedModelInfo('{"modelId":"gpt-3.5-turbo","provider":"OneHub"}');
		}
	}

	function storeSelectedModelInfo(selectedModelValue) {
		try {
			var selectedModel = JSON.parse(selectedModelValue);
			if (selectedModel && selectedModel.provider && selectedModel.modelId) {
				selectedModelInfo = {
					modelId: selectedModel.modelId,
					provider: selectedModel.provider,
					modelEndpoint: settings[selectedModel.provider]?.modelEndpoint,
					apiKey: settings[selectedModel.provider]?.apiKey,
					oneHubToken: settings[selectedModel.provider]?.oneHubToken,
				};
				localStorage.setItem(
					'selectedModelInfo',
					JSON.stringify(selectedModelInfo)
				);
			}
		} catch (error) {
			console.error('Error parsing selected model value:', error);
		}
	}

	function restoreSelectedModelInfo() {
		var storedModelInfo = localStorage.getItem('selectedModelInfo');
		if (storedModelInfo) {
			try {
				selectedModelInfo = JSON.parse(storedModelInfo);
				updateSingleModelOptions(); // 确保选项已经填充
				setSelectedModel(); // 然后设置选中的模型
			} catch (error) {
				console.error('Error parsing stored selected model info:', error);
				$('#singleModel').val('builtin').trigger('change');
			}
		} else {
			$('#singleModel').val('builtin').trigger('change');
		}
	}

	$('#serviceProvider').change(function () {
		var provider = $(this).val();
		updateProviderSettings(provider);
	});

	$('#singleModel').change(function () {
		storeSelectedModelInfo($(this).val());
	});

	$('#modelEndpoint, #apiKey, #oneHubToken, #model').on('change', function () {
		var provider = $('#serviceProvider').val();
		saveProviderSettings(provider);
		updateSingleModelOptions();
	});

	$('#oneHubToken').on('input', function () {
		var provider = $('#serviceProvider').val();
		saveProviderSettings(provider);
		updateModelSelect(settings[provider].models || []);
	});

	initSelect2();
	var defaultProvider = $('#serviceProvider').val();
	updateProviderSettings(defaultProvider); // 加载服务商设置
	restoreSelectedModelInfo(); // 恢复选中的模型信息
});
